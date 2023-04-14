package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
)

type Promotion struct {
	ID             string `json:"id"`
	Price          string `json:"price"`
	ExpirationDate string `json:"expiration_date"`
}

type PromotionsMap struct {
	sync.RWMutex
	m map[string]Promotion
}

func (p *PromotionsMap) Add(id string, promotion Promotion) {
	p.Lock()
	defer p.Unlock()
	p.m[id] = promotion
}

func (p *PromotionsMap) Get(id string) (Promotion, error) {
	p.RLock()
	defer p.RUnlock()
	promotion, ok := p.m[id]
	if !ok {
		return Promotion{}, fmt.Errorf("Promotion with ID %s not found", id)
	}
	return promotion, nil
}

func parsePromotionsCSVFile(file io.Reader) (map[string]Promotion, error) {
	r := csv.NewReader(file)
	records, err := r.ReadAll()
	if err != nil {
		return nil, err
	}
	promotions := make(map[string]Promotion)
	for _, record := range records {
		promotion := Promotion{
			ID:             record[0],
			Price:          record[1],
			ExpirationDate: record[2],
		}
		promotions[promotion.ID] = promotion
	}
	return promotions, nil
}

func main() {
	// Parse CSV file and store promotions in memory
	promotionsFile, err := os.Open("promotions.csv")
	if err != nil {
		log.Fatal(err)
	}
	promotions, err := parsePromotionsCSVFile(promotionsFile)
	if err != nil {
		log.Fatal(err)
	}
	promotionsMap := PromotionsMap{m: promotions}

	// HTTP endpoint to retrieve a promotion by ID
	http.HandleFunc("/promotions/", func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.Atoi(r.URL.Path[len("/promotions/"):])
		if err != nil {
			http.Error(w, "Invalid promotion ID", http.StatusBadRequest)
			return
		}
		promotion, err := promotionsMap.Get(strconv.Itoa(id))
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		promotionJSON, err := json.Marshal(promotion)
		if err != nil {
			http.Error(w, "Error encoding promotion to JSON", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(promotionJSON)
	})

	// Start HTTP server
	log.Println("Starting HTTP server on port 1321")
	log.Fatal(http.ListenAndServe(":1321", nil))
}
