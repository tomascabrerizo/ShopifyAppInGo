package main

type PackageItem struct {
	ProductID string
	Quantity int
}

func (app *Application) CalculatePackageVolumen(shop string, items []PackageItem) (float64, error) {
	var totalVolumen float64 = 0
	for _, item := range items {
		dim, err := app.GetProductDimensions(shop, item.ProductID)
		if err != nil {
			return 0, err
		}
		volumen := dim.Width * dim.Height * dim.Length
		totalVolumen += volumen * float64(item.Quantity)
	}
	return totalVolumen, nil
}
