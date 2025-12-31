package main

type PackageItem struct {
	ProductID string
	Quantity int
}

func (app *Application) calculatePackageVolumen(shop string, items []PackageItem) (float64, error) {
	token, err := app.db.GetAccessToken(shop)
	if err!= nil {
		return 0, err
	}

	var totalVolumen float64 = 0
	for _, item := range items {
		dim, err := app.shopApi.GetProductDimensions(shop, token.Access, item.ProductID)
		if err != nil {
			return 0, err
		}
		volumen := dim.Width * dim.Height * dim.Length
		totalVolumen += volumen * float64(item.Quantity)
	}
	return totalVolumen, nil
}
