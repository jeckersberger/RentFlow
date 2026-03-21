package application

import (
	"fmt"
	"math"
)

type PriceResult struct {
	Days                  int     `json:"days"`
	DailyRate             float64 `json:"daily_rate"`
	BasePrice             float64 `json:"base_price"`
	VolumeDiscount        float64 `json:"volume_discount"`
	CustomDiscount        float64 `json:"custom_discount"`
	Subtotal              float64 `json:"subtotal"`
	VAT                   float64 `json:"vat"`
	Total                 float64 `json:"total"`
	VolumeDiscountPercent float64 `json:"volume_discount_percent"`
	CustomDiscountPercent float64 `json:"custom_discount_percent"`
	VATPercent            float64 `json:"vat_percent"`
}

type PriceEngine struct {
	vatRate float64 // Default: 0.19 (19%)
}

func NewPriceEngine(vatRate float64) *PriceEngine {
	if vatRate == 0 {
		vatRate = 0.19 // Default German VAT
	}
	return &PriceEngine{
		vatRate: vatRate,
	}
}

// CalculatePrice calculates the rental price with volume discounts and VAT.
// dailyRate: daily rental rate
// days: number of rental days
// customDiscount: custom discount (0-1), e.g. 0.1 for 10% discount
func (pe *PriceEngine) CalculatePrice(dailyRate float64, days int, customDiscount float64) (*PriceResult, error) {
	if dailyRate < 0 {
		return nil, fmt.Errorf("daily rate cannot be negative")
	}
	if days <= 0 {
		return nil, fmt.Errorf("days must be greater than 0")
	}
	if customDiscount < 0 || customDiscount > 1 {
		return nil, fmt.Errorf("custom discount must be between 0 and 1")
	}

	// Calculate base price
	basePrice := dailyRate * float64(days)

	// Determine volume discount tier
	volumeDiscountPercent := 0.0
	if days >= 30 {
		volumeDiscountPercent = 0.20 // 20% for >= 30 days
	} else if days >= 14 {
		volumeDiscountPercent = 0.15 // 15% for >= 14 days
	} else if days >= 7 {
		volumeDiscountPercent = 0.10 // 10% for >= 7 days
	} else if days >= 3 {
		volumeDiscountPercent = 0.05 // 5% for >= 3 days
	}

	// Calculate discounts
	volumeDiscount := basePrice * volumeDiscountPercent
	customDiscountAmount := (basePrice - volumeDiscount) * customDiscount

	// Calculate subtotal after discounts
	subtotal := basePrice - volumeDiscount - customDiscountAmount

	// Calculate VAT
	vat := subtotal * pe.vatRate

	// Calculate total
	total := subtotal + vat

	return &PriceResult{
		Days:                  days,
		DailyRate:             dailyRate,
		BasePrice:             math.Round(basePrice*100) / 100,
		VolumeDiscount:        math.Round(volumeDiscount*100) / 100,
		CustomDiscount:        math.Round(customDiscountAmount*100) / 100,
		Subtotal:              math.Round(subtotal*100) / 100,
		VAT:                   math.Round(vat*100) / 100,
		Total:                 math.Round(total*100) / 100,
		VolumeDiscountPercent: volumeDiscountPercent,
		CustomDiscountPercent: customDiscount,
		VATPercent:            pe.vatRate,
	}, nil
}
