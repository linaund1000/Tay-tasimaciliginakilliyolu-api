package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/rs/cors"
	"go.uber.org/zap"
	"golang.org/x/time/rate"
)

type Truck struct {
	ID       string    `json:"id"`
	Location string    `json:"location"`
	Status   string    `json:"status"` // Available, In Transit, Maintenance
	LastUpdated time.Time `json:"last_updated"`
}

type Warehouse struct {
	ID       string         `json:"id"`
	Location string         `json:"location"`
	Capacity int            `json:"capacity"`
	Inventory map[string]int `json:"inventory"` // Item ID to quantity
}

type Mission struct {
	ID          string    `json:"id"`
	TruckID     string    `json:"truck_id"`
	FromID      string    `json:"from_id"`
	ToID        string    `json:"to_id"`
	Status      string    `json:"status"` // Created, In Progress, Completed, Failed
	ItemID      string    `json:"item_id"`
	Quantity    int       `json:"quantity"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type LogisticsSystem struct {
	Trucks     map[string]*Truck
	Warehouses map[string]*Warehouse
	Missions   map[string]*Mission
	mu         sync.RWMutex
	logger     *zap.Logger
	limiter    *rate.Limiter
}

func NewLogisticsSystem() *LogisticsSystem {
	logger, _ := zap.NewProduction()
	return &LogisticsSystem{
		Trucks:     make(map[string]*Truck),
		Warehouses: make(map[string]*Warehouse),
		Missions:   make(map[string]*Mission),
		logger:     logger,
		limiter:    rate.NewLimiter(rate.Every(time.Second), 100), // 100 requests per second
	}
}

func (ls *LogisticsSystem) rateLimitMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !ls.limiter.Allow() {
			http.Error(w, "Too many requests", http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	}
}

func (ls *LogisticsSystem) AddTruck(w http.ResponseWriter, r *http.Request) {
	var truck Truck
	if err := json.NewDecoder(r.Body).Decode(&truck); err != nil {
		ls.logger.Error("Failed to decode truck", zap.Error(err))
		http.Error(w, "Invalid truck data", http.StatusBadRequest)
		return
	}

	if truck.ID == "" {
		truck.ID = uuid.New().String()
	}
	truck.LastUpdated = time.Now()

	ls.mu.Lock()
	ls.Trucks[truck.ID] = &truck
	ls.mu.Unlock()

	ls.logger.Info("Added new truck", zap.String("id", truck.ID))
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(truck)
}

func (ls *LogisticsSystem) GetTruck(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	truckID := vars["id"]

	ls.mu.RLock()
	truck, exists := ls.Trucks[truckID]
	ls.mu.RUnlock()

	if !exists {
		http.Error(w, "Truck not found", http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(truck)
}

func (ls *LogisticsSystem) UpdateTruck(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	truckID := vars["id"]

	var updatedTruck Truck
	if err := json.NewDecoder(r.Body).Decode(&updatedTruck); err != nil {
		ls.logger.Error("Failed to decode truck update", zap.Error(err))
		http.Error(w, "Invalid truck data", http.StatusBadRequest)
		return
	}

	ls.mu.Lock()
	truck, exists := ls.Trucks[truckID]
	if !exists {
		ls.mu.Unlock()
		http.Error(w, "Truck not found", http.StatusNotFound)
		return
	}

	truck.Location = updatedTruck.Location
	truck.Status = updatedTruck.Status
	truck.LastUpdated = time.Now()
	ls.mu.Unlock()

	ls.logger.Info("Updated truck", zap.String("id", truckID))
	json.NewEncoder(w).Encode(truck)
}

func (ls *LogisticsSystem) AddWarehouse(w http.ResponseWriter, r *http.Request) {
	var warehouse Warehouse
	if err := json.NewDecoder(r.Body).Decode(&warehouse); err != nil {
		ls.logger.Error("Failed to decode warehouse", zap.Error(err))
		http.Error(w, "Invalid warehouse data", http.StatusBadRequest)
		return
	}

	if warehouse.ID == "" {
		warehouse.ID = uuid.New().String()
	}

	ls.mu.Lock()
	ls.Warehouses[warehouse.ID] = &warehouse
	ls.mu.Unlock()

	ls.logger.Info("Added new warehouse", zap.String("id", warehouse.ID))
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(warehouse)
}

func (ls *LogisticsSystem) GetWarehouse(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	warehouseID := vars["id"]

	ls.mu.RLock()
	warehouse, exists := ls.Warehouses[warehouseID]
	ls.mu.RUnlock()

	if !exists {
		http.Error(w, "Warehouse not found", http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(warehouse)
}

func (ls *LogisticsSystem) UpdateWarehouseInventory(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	warehouseID := vars["id"]

	var inventoryUpdate map[string]int
	if err := json.NewDecoder(r.Body).Decode(&inventoryUpdate); err != nil {
		ls.logger.Error("Failed to decode inventory update", zap.Error(err))
		http.Error(w, "Invalid inventory data", http.StatusBadRequest)
		return
	}

	ls.mu.Lock()
	warehouse, exists := ls.Warehouses[warehouseID]
	if !exists {
		ls.mu.Unlock()
		http.Error(w, "Warehouse not found", http.StatusNotFound)
		return
	}

	for itemID, quantity := range inventoryUpdate {
		warehouse.Inventory[itemID] += quantity
		if warehouse.Inventory[itemID] < 0 {
			warehouse.Inventory[itemID] = 0
		}
	}
	ls.mu.Unlock()

	ls.logger.Info("Updated warehouse inventory", zap.String("id", warehouseID))
	json.NewEncoder(w).Encode(warehouse)
}

func (ls *LogisticsSystem) CreateMission(w http.ResponseWriter, r *http.Request) {
	var mission Mission
	if err := json.NewDecoder(r.Body).Decode(&mission); err != nil {
		ls.logger.Error("Failed to decode mission", zap.Error(err))
		http.Error(w, "Invalid mission data", http.StatusBadRequest)
		return
	}

	ls.mu.Lock()
	defer ls.mu.Unlock()

	if _, ok := ls.Trucks[mission.TruckID]; !ok {
		http.Error(w, "Truck not found", http.StatusNotFound)
		return
	}
	if _, ok := ls.Warehouses[mission.FromID]; !ok {
		http.Error(w, "Source warehouse not found", http.StatusNotFound)
		return
	}
	if _, ok := ls.Warehouses[mission.ToID]; !ok {
		http.Error(w, "Destination warehouse not found", http.StatusNotFound)
		return
	}

	sourceWarehouse := ls.Warehouses[mission.FromID]
	if sourceWarehouse.Inventory[mission.ItemID] < mission.Quantity {
		http.Error(w, "Insufficient inventory", http.StatusBadRequest)
		return
	}

	sourceWarehouse.Inventory[mission.ItemID] -= mission.Quantity
	ls.Trucks[mission.TruckID].Status = "In Transit"

	mission.ID = uuid.New().String()
	mission.Status = "Created"
	mission.CreatedAt = time.Now()
	mission.UpdatedAt = time.Now()
	ls.Missions[mission.ID] = &mission

	ls.logger.Info("Created new mission", zap.String("id", mission.ID))
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(mission)
}

func (ls *LogisticsSystem) GetMission(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	missionID := vars["id"]

	ls.mu.RLock()
	mission, exists := ls.Missions[missionID]
	ls.mu.RUnlock()

	if !exists {
		http.Error(w, "Mission not found", http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(mission)
}

func (ls *LogisticsSystem) UpdateMissionStatus(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	missionID := vars["id"]

	var update struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&update); err != nil {
		ls.logger.Error("Failed to decode status update", zap.Error(err))
		http.Error(w, "Invalid status data", http.StatusBadRequest)
		return
	}

	ls.mu.Lock()
	defer ls.mu.Unlock()

	mission, ok := ls.Missions[missionID]
	if !ok {
		http.Error(w, "Mission not found", http.StatusNotFound)
		return
	}

	mission.Status = update.Status
	mission.UpdatedAt = time.Now()

	if update.Status == "Completed" {
		destWarehouse := ls.Warehouses[mission.ToID]
		destWarehouse.Inventory[mission.ItemID] += mission.Quantity
		ls.Trucks[mission.TruckID].Status = "Available"
	}

	ls.logger.Info("Updated mission status", zap.String("id", missionID), zap.String("status", update.Status))
	json.NewEncoder(w).Encode(mission)
}

func (ls *LogisticsSystem) GetMissions(w http.ResponseWriter, r *http.Request) {
	ls.mu.RLock()
	missions := make([]*Mission, 0, len(ls.Missions))
	for _, mission := range ls.Missions {
		missions = append(missions, mission)
	}
	ls.mu.RUnlock()

	json.NewEncoder(w).Encode(missions)
}


func main() {
	ls := NewLogisticsSystem()
	defer ls.logger.Sync()

	r := mux.NewRouter()

	// Truck routes
	r.HandleFunc("/truck", ls.rateLimitMiddleware(ls.AddTruck)).Methods("POST")
	r.HandleFunc("/truck/{id}", ls.rateLimitMiddleware(ls.GetTruck)).Methods("GET")
	r.HandleFunc("/truck/{id}", ls.rateLimitMiddleware(ls.UpdateTruck)).Methods("PUT")

	// Warehouse routes
	r.HandleFunc("/warehouse", ls.rateLimitMiddleware(ls.AddWarehouse)).Methods("POST")
	r.HandleFunc("/warehouse/{id}", ls.rateLimitMiddleware(ls.GetWarehouse)).Methods("GET")
	r.HandleFunc("/warehouse/{id}/inventory", ls.rateLimitMiddleware(ls.UpdateWarehouseInventory)).Methods("PUT")

	// Mission routes
	r.HandleFunc("/mission", ls.rateLimitMiddleware(ls.CreateMission)).Methods("POST")
	r.HandleFunc("/mission/{id}", ls.rateLimitMiddleware(ls.GetMission)).Methods("GET")
	r.HandleFunc("/mission/{id}/status", ls.rateLimitMiddleware(ls.UpdateMissionStatus)).Methods("PUT")
	r.HandleFunc("/missions", ls.rateLimitMiddleware(ls.GetMissions)).Methods("GET")

	c := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE"},
		AllowedHeaders: []string{"Content-Type"},
	})

	handler := c.Handler(r)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	server := &http.Server{
		Addr:         ":" + port,
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	ls.logger.Info("Server is running", zap.String("port", port))
	log.Fatal(server.ListenAndServe())

}