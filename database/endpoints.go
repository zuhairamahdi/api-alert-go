package database

import (
	"api-monitor/models"
	"log"
	"time"
)

// LoadEndpoints loads all endpoints from the database
func LoadEndpoints() ([]models.Endpoint, error) {
	var dbEndpoints []Endpoint
	if err := DB.Find(&dbEndpoints).Error; err != nil {
		return nil, err
	}

	endpoints := make([]models.Endpoint, len(dbEndpoints))
	for i, dbEndpoint := range dbEndpoints {
		endpoints[i] = dbEndpoint.ToModel()
	}

	return endpoints, nil
}

// CreateEndpoint creates a new endpoint in the database
func CreateEndpoint(endpoint *models.Endpoint) error {
	dbEndpoint := FromModel(*endpoint)

	if err := DB.Create(&dbEndpoint).Error; err != nil {
		return err
	}

	endpoint.ID = int(dbEndpoint.ID)
	return nil
}

// UpdateEndpoint updates an existing endpoint in the database
func UpdateEndpoint(endpoint *models.Endpoint) error {
	dbEndpoint := FromModel(*endpoint)
	dbEndpoint.ID = uint(endpoint.ID)

	return DB.Model(&Endpoint{}).Where("id = ?", endpoint.ID).Updates(dbEndpoint).Error
}

// DeleteEndpoint deletes an endpoint from the database
func DeleteEndpoint(id int) error {
	return DB.Delete(&Endpoint{}, id).Error
}

// UpdateEndpointStatus updates the status and last checked time of an endpoint
func UpdateEndpointStatus(id int, status string) error {
	now := time.Now()

	// Try up to 3 times
	maxRetries := 3
	for i := 0; i < maxRetries; i++ {
		err := DB.Model(&Endpoint{}).Where("id = ?", id).Updates(map[string]interface{}{
			"status":       status,
			"last_checked": now,
		}).Error

		if err == nil {
			return nil
		}

		if i < maxRetries-1 {
			log.Printf("Retry %d/%d for updating endpoint status: %v", i+1, maxRetries, err)
			time.Sleep(1 * time.Second)
			continue
		}

		return err
	}

	return nil
}
