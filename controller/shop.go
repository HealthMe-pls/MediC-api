package controller

import (
	"fmt"
	"strconv"

	// "github.com/HealthMe-pls/medic-go-api/controller"
	"github.com/HealthMe-pls/medic-go-api/model"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func Shop(db *gorm.DB, c *fiber.Ctx) error {
	var entrepreneur []model.Entrepreneur
	db.Find(&entrepreneur)
	return c.JSON(entrepreneur)
}
func GetShopByID(db *gorm.DB, c *fiber.Ctx) error {
	id := c.Params("id")
	var shop model.Shop
	if err := db.First(&shop, id).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error":   "Shop not found",
			"details": err.Error(),
		})
	}
	// db.First(&shop, id)
	return c.JSON(shop)
}
func GetShops(db *gorm.DB, c *fiber.Ctx) error {
	var shops []model.Shop
	if err := db.Find(&shops).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Failed to retrieve shop categories",
			"details": err.Error(),
		})
	}
	return c.JSON(shops)
}

func GetShopDetail(db *gorm.DB, c *fiber.Ctx) error {
	var shops []model.Shop

	// Fetch basic shop details with Entrepreneur and ShopCategory preloaded
	if err := db.Preload("Entrepreneur").Preload("ShopCategory").Find(&shops).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Failed to retrieve shops",
			"details": err.Error(),
		})
	}

	// Construct the detailed response
	var shopResponses []fiber.Map
	for _, shop := range shops {
		shopID := shop.ID

		// Fetch shop open dates
		shopOpenDates, err := getShopOpenDates(db, shopID)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":   "Failed to retrieve shop open dates",
				"details": err.Error(),
			})
		}

		// Fetch shop menus
		shopMenus, err := getShopMenus(db, shopID)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":   "Failed to retrieve shop menus",
				"details": err.Error(),
			})
		}

		// Fetch social media
		socialMedias, err := getSocialMedia(db, shopID)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":   "Failed to retrieve social media entries",
				"details": err.Error(),
			})
		}

		// Fetch all photos related to the shop by shopID
		shopPhotos, err := getPhotosByShopID(db, shopID)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":   "Failed to retrieve photos",
				"details": err.Error(),
			})
		}

		// Construct the shop response
		shopResponses = append(shopResponses, fiber.Map{
			"shop_id":         shop.ID,
			"name":            shop.Name,
			"entrepreneur_id": shop.Entrepreneur.ID,
			"entrepreneur":    shop.Entrepreneur.Title + " " + shop.Entrepreneur.FirstName + " " + shop.Entrepreneur.MiddleName + " " + shop.Entrepreneur.LastName,
			"category_id":     shop.ShopCategory.ID,
			"category":        shop.ShopCategory.Name,
			"open_status":     shop.OpenStatus,
			"description":     shop.Description,
			"photos":          shopPhotos, // Updated to include all photos related to the shop
			"shop_open_dates": shopOpenDates,
			"menus":           shopMenus,
			"social_media":    socialMedias,
		})
	}

	return c.JSON(shopResponses)
}

func stringToUint(shopID string) (uint, error) {
	// Log shopID to verify it
	fmt.Println("Received shopID:", shopID)

	// Try to convert the string to uint
	id, err := strconv.ParseUint(shopID, 10, 32)
	if err != nil {
		// Log error for debugging
		fmt.Println("Error parsing shopID:", err)
		return 0, err
	}

	return uint(id), nil
}

func GetShopDetailByID(db *gorm.DB, c *fiber.Ctx) error {
	shopID := c.Params("id") // shop_id is still a string
	fmt.Println("shopID from URL parameter:", shopID)

	// Convert the string shopID to uint
	shopIDUint, err := stringToUint(shopID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid shop ID",
		})
	}

	// Fetch a single shop by ID with Entrepreneur and ShopCategory preloaded
	var shop model.Shop
	if err := db.Preload("Entrepreneur").Preload("ShopCategory").First(&shop, "id = ?", shopIDUint).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Failed to retrieve shop",
			"details": err.Error(),
		})
	}

	// Fetch related data using helper functions
	shopOpenDates, err := getShopOpenDates(db, shopIDUint) // Fetch shop open dates
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Failed to retrieve shop open dates",
			"details": err.Error(),
		})
	}

	shopMenus, err := getShopMenus(db, shopIDUint) // Fetch shop menus
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Failed to retrieve shop menus",
			"details": err.Error(),
		})
	}

	socialMedias, err := getSocialMedia(db, shopIDUint) // Fetch social media links
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Failed to retrieve social media entries",
			"details": err.Error(),
		})
	}

	// Fetch all photos related to the shop using shopID
	shopPhotos, err := getPhotosByShopID(db, shopIDUint)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Failed to retrieve shop photos",
			"details": err.Error(),
		})
	}

	// Construct the shop response
	shopResponse := fiber.Map{
		"shop_id":         shop.ID,
		"name":            shop.Name,
		"entrepreneur_id": shop.Entrepreneur.ID,
		"entrepreneur":    shop.Entrepreneur.Title + " " + shop.Entrepreneur.FirstName + " " + shop.Entrepreneur.MiddleName + " " + shop.Entrepreneur.LastName,
		"category_id":     shop.ShopCategory.ID,
		"category":        shop.ShopCategory.Name,
		"open_status":     shop.OpenStatus,
		"description":     shop.Description,
		"photos":          shopPhotos, // Include all photos related to the shop
		"shop_open_dates": shopOpenDates,
		"menus":           shopMenus,
		"social_media":    socialMedias,
	}

	return c.JSON(shopResponse)
}

func getShopOpenDates(db *gorm.DB, shopID uint) ([]fiber.Map, error) {
	var shopOpenDates []model.ShopOpenDate
	if err := db.Where("shop_id = ?", shopID).Find(&shopOpenDates).Error; err != nil {
		return nil, err
	}

	var result []fiber.Map
	for _, date := range shopOpenDates {
		result = append(result, fiber.Map{
			"id":         date.ID,
			"start_time": date.StartTime,
			"end_time":   date.EndTime,
		})
	}
	return result, nil
}

// Helper function to fetch shop menus
func getShopMenus(db *gorm.DB, shopID uint) ([]fiber.Map, error) {
	var shopMenus []model.ShopMenu
	if err := db.Where("shop_id = ?", shopID).Find(&shopMenus).Error; err != nil {
		return nil, err
	}

	var result []fiber.Map
	for _, menu := range shopMenus {
		// Fetch all photos related to the menu by MenuID
		menuPhotos, err := getPhotosByMenuID(db, menu.ID)
		if err != nil {
			return nil, err
		}
		result = append(result, fiber.Map{
			"id":                  menu.ID,
			"product_name":        menu.ProductName,
			"product_description": menu.ProductDescription,
			"price":               menu.Price,
			"photos":              menuPhotos, // Include all photos related to the menu
		})
	}
	return result, nil
}

// Helper function to fetch social media
func getSocialMedia(db *gorm.DB, shopID uint) ([]fiber.Map, error) {
	var socialMedias []model.SocialMedia
	if err := db.Where("shop_id = ?", shopID).Find(&socialMedias).Error; err != nil {
		return nil, err
	}

	var result []fiber.Map
	for _, social := range socialMedias {
		result = append(result, fiber.Map{
			"id":       social.ID,
			"platform": social.Platform,
			"link":     social.Link,
		})
	}
	return result, nil
}

func getPhotosByShopID(db *gorm.DB, shopID uint) ([]fiber.Map, error) {
	var photos []model.Photo
	if err := db.Where("shop_id = ?", shopID).Find(&photos).Error; err != nil {
		return nil, err
	}

	var result []fiber.Map
	for _, photo := range photos {
		result = append(result, fiber.Map{
			"photo_id": photo.ID,
			"pathfile": photo.PathFile,
		})
	}
	return result, nil
}

func getPhotosByMenuID(db *gorm.DB, menuID uint) ([]fiber.Map, error) {
	var photos []model.Photo
	if err := db.Where("menu_id = ?", menuID).Find(&photos).Error; err != nil {
		return nil, err
	}

	var result []fiber.Map
	for _, photo := range photos {
		result = append(result, fiber.Map{
			"photo_id": photo.ID,
			"pathfile": photo.PathFile,
		})
	}
	return result, nil
}

func CreateShop(db *gorm.DB, c *fiber.Ctx) error {
	// Parse the request body into the Shop struct
	shop := new(model.Shop)
	if err := c.BodyParser(shop); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Failed to parse request body",
			"details": err.Error(),
		})
	}

	// Check if the ShopCategory exists by its ID
	var shopCategory model.ShopCategory
	if err := db.First(&shopCategory, shop.ShopCategoryID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error":   "ShopCategory not found",
			"details": err.Error(),
		})
	}

	// Save the Shop to the database
	if result := db.Create(&shop); result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create shop",
		})
	}

	// Return the created Shop as a JSON response
	return c.Status(fiber.StatusCreated).JSON(shop)
}
func UpdateShop(db *gorm.DB, c *fiber.Ctx) error {
	// Get the shop ID parameter from the URL
	id := c.Params("id")
	var shop model.Shop

	// Find the shop by ID
	if err := db.First(&shop, id).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Shop not found",
			"details": err.Error(),
		})
		// return c.Status(fiber.StatusNotFound).SendString("Shop not found")
	}

	// Parse the updated details from the request body
	if err := c.BodyParser(&shop); err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("Failed to parse request body")
	}

	// Save the updated shop details to the database
	if result := db.Save(&shop); result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Failed to update shop")
	}

	// Return the updated shop as a JSON response
	return c.JSON(shop)
}
func DeleteShop(db *gorm.DB, c *fiber.Ctx) error {
	// Get the shop ID parameter from the URL
	id := c.Params("id")

	// Delete the shop from the database by its ID
	if result := db.Delete(&model.Shop{}, id); result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Failed to delete shop")
	}

	// Return success message
	return c.SendString("Shop successfully deleted")
}

func GetShopsByCategory(db *gorm.DB, c *fiber.Ctx) error {
	// Get the ShopCategoryID parameter from the URL
	shopCategoryID := c.Params("shop_category_id")

	var shops []model.Shop

	// Query shops by the ShopCategoryID
	if err := db.Where("shop_category_id = ?", shopCategoryID).Find(&shops).Error; err != nil {
		return c.Status(fiber.StatusNotFound).SendString("No shops found for this category")
	}

	// Return the shops as a JSON response
	return c.JSON(shops)
}

// -----social media
// CreateSocialMedia creates a new SocialMedia entry
func CreateSocialMedia(db *gorm.DB, c *fiber.Ctx) error {
	var socialMedia model.SocialMedia
	if err := c.BodyParser(&socialMedia); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request payload",
		})
	}
	if err := db.Create(&socialMedia).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create social media",
		})
	}
	return c.Status(fiber.StatusCreated).JSON(socialMedia)
}

// GetSocialMedia retrieves a SocialMedia entry by ID
func GetSocialMedia(db *gorm.DB, c *fiber.Ctx) error {
	id := c.Params("id")
	var socialMedia model.SocialMedia
	// if err := db.First(&socialMedia, id).Error; err != nil {
	// 	return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
	// 		"error": "Social media not found",
	// 	})
	// }
	db.First(&socialMedia, id)
	return c.JSON(socialMedia)
}

// GetSocialMediaByShopID retrieves SocialMedia entries by Shop ID
func GetSocialMediaByShopID(db *gorm.DB, c *fiber.Ctx) error {
	shopID := c.Params("shop_id")
	var socialMedias []map[string]interface{}

	// Query specific fields
	if err := db.Model(&model.SocialMedia{}).
		Select("id, platform, link").
		Where("shop_id = ?", shopID).
		Find(&socialMedias).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to retrieve social media entries",
		})
	}

	return c.JSON(socialMedias)
}

// UpdateSocialMedia updates a SocialMedia entry by ID
func UpdateSocialMedia(db *gorm.DB, c *fiber.Ctx) error {
	id := c.Params("id")
	var socialMedia model.SocialMedia
	if err := db.First(&socialMedia, id).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Social media not found",
		})
	}

	if err := c.BodyParser(&socialMedia); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request payload",
		})
	}

	if err := db.Save(&socialMedia).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update social media",
		})
	}
	return c.JSON(socialMedia)
}

// DeleteSocialMedia deletes a SocialMedia entry by ID
func DeleteSocialMedia(db *gorm.DB, c *fiber.Ctx) error {
	id := c.Params("id")
	if err := db.Delete(&model.SocialMedia{}, id).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to delete social media",
		})
	}
	return c.SendStatus(fiber.StatusNoContent)
}

// shop menu
// CreateShopMenu creates a new ShopMenu entry
func CreateShopMenu(db *gorm.DB, c *fiber.Ctx) error {
	var shopMenu model.ShopMenu
	if err := c.BodyParser(&shopMenu); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request payload",
		})
	}
	if err := db.Create(&shopMenu).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create shop menu",
		})
	}
	return c.Status(fiber.StatusCreated).JSON(shopMenu)
}

// GetShopMenu retrieves a ShopMenu entry by ID
func GetShopMenu(db *gorm.DB, c *fiber.Ctx) error {
	id := c.Params("id")
	var shopMenu model.ShopMenu
	// if err := db.First(&shopMenu, id).Error; err != nil {
	// 	return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
	// 		"error": "Shop menu not found",
	// 	})
	// }
	db.First(&shopMenu, id)
	return c.JSON(shopMenu)
}

// GetShopMenuByShopID retrieves ShopMenu entries by Shop ID
func GetShopMenuByShopID(db *gorm.DB, c *fiber.Ctx) error {
	shopID := c.Params("shop_id")
	var shopMenus []map[string]interface{}

	// Query specific fields
	if err := db.Model(&model.ShopMenu{}).
		Select("id, product_description, price, product_name, photo").
		Where("shop_id = ?", shopID).
		Find(&shopMenus).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to retrieve shop menus",
		})
	}

	return c.JSON(shopMenus)
}

// UpdateShopMenu updates a ShopMenu entry by ID
func UpdateShopMenu(db *gorm.DB, c *fiber.Ctx) error {
	id := c.Params("id")
	var shopMenu model.ShopMenu
	if err := db.First(&shopMenu, id).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Shop menu not found",
		})
	}

	if err := c.BodyParser(&shopMenu); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request payload",
		})
	}

	if err := db.Save(&shopMenu).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update shop menu",
		})
	}
	return c.JSON(shopMenu)
}

// DeleteShopMenu deletes a ShopMenu entry by ID
func DeleteShopMenu(db *gorm.DB, c *fiber.Ctx) error {
	id := c.Params("id")
	if err := db.Delete(&model.ShopMenu{}, id).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to delete shop menu",
		})
	}
	return c.SendStatus(fiber.StatusNoContent)
}

//-----shop category

func CreateShopCategory(db *gorm.DB, c *fiber.Ctx) error {
	// Parse the request body into the ShopCategory struct
	shopCategory := new(model.ShopCategory)
	if err := c.BodyParser(shopCategory); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Failed to parse request body",
		})
	}

	// Create the ShopCategory record in the database
	if err := db.Create(shopCategory).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create ShopCategory",
		})
	}

	// Return the newly created ShopCategory as JSON response
	return c.Status(fiber.StatusCreated).JSON(shopCategory)
}

func GetShopCategories(db *gorm.DB, c *fiber.Ctx) error {
	var categories []model.ShopCategory
	if err := db.Find(&categories).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to retrieve shop categories",
		})
	}
	return c.JSON(categories)
}
func GetShopCategoryByID(db *gorm.DB, c *fiber.Ctx) error {
	id := c.Params("id")
	var category model.ShopCategory
	// if err := db.First(&category, id).Error; err != nil {
	// 	return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
	// 		"error": "Shop category not found",
	// 	})
	// }
	db.First(&category, id)
	return c.JSON(category)
}

func UpdateShopCategory(db *gorm.DB, c *fiber.Ctx) error {
	// Get the shop category ID from the URL parameters
	shopCategoryID := c.Params("id")

	// Convert the ID to uint
	id, err := strconv.Atoi(shopCategoryID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("Invalid ShopCategory ID format")
	}

	// Fetch the existing shop category
	var shopCategory model.ShopCategory
	if err := db.First(&shopCategory, id).Error; err != nil {
		return c.Status(fiber.StatusNotFound).SendString("ShopCategory not found")
	}

	// Parse the request body for updates
	if err := c.BodyParser(&shopCategory); err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("Failed to parse request body")
	}

	// Save the updated shop category to the database
	if err := db.Save(&shopCategory).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Failed to update ShopCategory")
	}

	// Return the updated shop category as JSON
	return c.Status(fiber.StatusOK).JSON(shopCategory)
}

func DeleteShopCategory(db *gorm.DB, c *fiber.Ctx) error {
	// Get the shop category ID from the URL parameters
	shopCategoryID := c.Params("id")

	// Convert the ID to uint
	id, err := strconv.Atoi(shopCategoryID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("Invalid ShopCategory ID format")
	}

	// Attempt to delete the shop category
	if err := db.Delete(&model.ShopCategory{}, id).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Failed to delete ShopCategory")
	}

	// Return success message
	return c.SendString("ShopCategory deleted successfully")
}
