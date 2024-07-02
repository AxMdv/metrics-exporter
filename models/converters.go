package Data

import (
	"fmt"
	"regexp"
	"strconv"
)

func convertIntArrayToString(input []int) (result string) {
	for _, value := range input {
		if result != "" {
			result = result + ","
		}
		result = result + fmt.Sprintf("%v", value)
	}
	return
}

func convertPriceArrayToStr(input []float32) (result []string) {
	for _, value := range input {
		result = append(result, fmt.Sprintf("%.2f", value))
	}
	if len(result) == 1 {
		result = append(result, "")
	}
	return
}

func (article *BOArticle) ToEslArticle() (eslArticle CustomESLArticle) {
	if article.Plu != "" {
		eslArticle.Plu = article.Plu
		eslArticle.EAN = article.EAN
		eslArticle.ItemName = regexp.MustCompile(`\s+`).ReplaceAllString(article.ItemName, " ")
		eslArticle.Description = article.Description
		eslArticle.Manufacturer = article.Manufacturer
		eslArticle.Stpr = article.Stpr
		eslArticle.Unit = article.Unit
		eslArticle.Price = convertPriceArrayToStr(article.Price)[0]
		eslArticle.OldPrice = fmt.Sprintf("%.2f", article.OldPrice)
		eslArticle.PromoPrice = convertPriceArrayToStr(article.Price)[1]
		eslArticle.PromoQty = convertIntArrayToString(article.PromoQty)
		eslArticle.PromoType = article.PromoType
		eslArticle.Ctm = article.Ctm
		eslArticle.TopType = article.TopType
		eslArticle.Ingredients = article.Ingredients
		eslArticle.Promo = strconv.FormatBool(article.Promo)
		eslArticle.ScaleNumber = strconv.Itoa(article.ScaleNumber)
		eslArticle.ItemPlace = article.ItemPlace
		eslArticle.StoreName = article.StoreName
		eslArticle.InAssortment = strconv.FormatBool(article.InAssortment)
		eslArticle.ProdDate = article.ProdDate
		eslArticle.Style = article.Style
		eslArticle.Location = article.Location
		eslArticle.Alcohol = article.Alcohol
	}
	return eslArticle
}
