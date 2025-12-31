package main

import (
	"unicode"

	"tomi/src/shopify"
)

type PackageItem struct {
	ProductID string
	Quantity int
}

func calculatePackageVolumen(api *shopify.Api, token string, shop string, items []PackageItem) (float64, error) {
	var totalVolumen float64 = 0
	for _, item := range items {
		dim, err := api.GetProductDimensions(shop, token, item.ProductID)
		if err != nil {
			return 0, err
		}
		volumen := dim.Width * dim.Height * dim.Length
		totalVolumen += volumen * float64(item.Quantity)
	}
	return totalVolumen, nil
}

func onlyDigits(s string) string {
	var b []rune
	for _, c := range s {
		if unicode.IsDigit(c) {
			b = append(b, c)
		}
	}
	return string(b)
}
