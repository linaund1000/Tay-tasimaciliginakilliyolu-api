Evet, bu kodun API tasarım belgesini Türkçe olarak yazabilirim. Aşağıda API tasarım dokümantasyonu bulunmaktadır:
Logistics System API Tasarım Dokümantasyonu

Bu dokümantasyon, Logistics System API'sinin kullanımını açıklamaktadır. API, kamyonların, depoların ve görevlerin yönetimini sağlar.
Endpointler
Kamyonlar (Trucks)
Kamyon Ekleme

    Endpoint: /truck
    Method: POST
    Description: Yeni bir kamyon ekler.
    Request Body:

{
  "id": "string", // Opsiyonel
  "location": "string",
  "status": "string" // Available, In Transit, Maintenance
}

Response:

    {
      "id": "string",
      "location": "string",
      "status": "string",
      "last_updated": "timestamp"
    }

Kamyon Alma

    Endpoint: /truck/{id}
    Method: GET
    Description: Belirtilen ID'ye sahip kamyonu getirir.
    Response:

    {
      "id": "string",
      "location": "string",
      "status": "string",
      "last_updated": "timestamp"
    }

Kamyon Güncelleme

    Endpoint: /truck/{id}
    Method: PUT
    Description: Belirtilen ID'ye sahip kamyonu günceller.
    Request Body:

{
  "location": "string",
  "status": "string"
}

Response:

    {
      "id": "string",
      "location": "string",
      "status": "string",
      "last_updated": "timestamp"
    }

Depolar (Warehouses)
Depo Ekleme

    Endpoint: /warehouse
    Method: POST
    Description: Yeni bir depo ekler.
    Request Body:

{
  "id": "string", // Opsiyonel
  "location": "string",
  "capacity": int,
  "inventory": {
    "item_id": int
  }
}

Response:

    {
      "id": "string",
      "location": "string",
      "capacity": int,
      "inventory": {
        "item_id": int
      }
    }

Depo Alma

    Endpoint: /warehouse/{id}
    Method: GET
    Description: Belirtilen ID'ye sahip depo getirir.
    Response:

    {
      "id": "string",
      "location": "string",
      "capacity": int,
      "inventory": {
        "item_id": int
      }
    }

Depo Envanteri Güncelleme

    Endpoint: /warehouse/{id}/inventory
    Method: PUT
    Description: Belirtilen ID'ye sahip depo envanterini günceller.
    Request Body:

{
  "item_id": int
}

Response:

    {
      "id": "string",
      "location": "string",
      "capacity": int,
      "inventory": {
        "item_id": int
      }
    }

Görevler (Missions)
Görev Oluşturma

    Endpoint: /mission
    Method: POST
    Description: Yeni bir görev oluşturur.
    Request Body:

{
  "truck_id": "string",
  "from_id": "string",
  "to_id": "string",
  "item_id": "string",
  "quantity": int
}

Response:

    {
      "id": "string",
      "truck_id": "string",
      "from_id": "string",
      "to_id": "string",
      "status": "string", // Created, In Progress, Completed, Failed
      "item_id": "string",
      "quantity": int,
      "created_at": "timestamp",
      "updated_at": "timestamp"
    }

Görev Alma

    Endpoint: /mission/{id}
    Method: GET
    Description: Belirtilen ID'ye sahip görevi getirir.
    Response:

    {
      "id": "string",
      "truck_id": "string",
      "from_id": "string",
      "to_id": "string",
      "status": "string",
      "item_id": "string",
      "quantity": int,
      "created_at": "timestamp",
      "updated_at": "timestamp"
    }

Görev Durumu Güncelleme

    Endpoint: /mission/{id}/status
    Method: PUT
    Description: Belirtilen ID'ye sahip görevin durumunu günceller.
    Request Body:

{
  "status": "string" // Created, In Progress, Completed, Failed
}

Response:

    {
      "id": "string",
      "truck_id": "string",
      "from_id": "string",
      "to_id": "string",
      "status": "string",
      "item_id": "string",
      "quantity": int,
      "created_at": "timestamp",
      "updated_at": "timestamp"
    }

Tüm Görevleri Alma

    Endpoint: /missions
    Method: GET
    Description: Tüm görevleri getirir.
    Response:

    [
      {
        "id": "string",
        "truck_id": "string",
        "from_id": "string",
        "to_id": "string",
        "status": "string",
        "item_id": "string",
        "quantity": int,
        "created_at": "timestamp",
        "updated_at": "timestamp"
      }
    ]

Hata Kodları

    400 Bad Request: Geçersiz istek verisi.
    404 Not Found: Kaynak bulunamadı.
    429 Too Many Requests: Çok fazla istek.

Rate Limiting

API, saniyede 100 istek ile sınırlıdır. Bu sınırı aşan istekler için 429 Too Many Requests hatası döndürülür.