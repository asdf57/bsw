package controller

import (
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/asdf57/bsw/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Controller struct {
	db *gorm.DB
}

func NewController(db *gorm.DB) *Controller {
	return &Controller{db: db}
}

func (ctrl *Controller) resolveUserIdFromName(name string) (*uint, error) {
	var dbEntry models.UserDBEntry

	if err := ctrl.db.Where("name = ?", name).First(&dbEntry).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Printf("user was not found: %v", err)
			return nil, err
		}

		log.Printf("unhandled error while searching for user record: %v", err)
		return nil, err
	}

	return &dbEntry.ID, nil
}

func (ctrl *Controller) getUsers() ([]models.UserDBEntry, error) {
	var users []models.UserDBEntry

	if err := ctrl.db.Find(&users).Error; err != nil {
		log.Printf("db error while fetching users: %v", err)
		return nil, err
	}

	return users, nil
}

func (ctrl *Controller) resolveUserNameFromId(id uint) (*string, error) {
	var dbEntry models.UserDBEntry

	if err := ctrl.db.First(&dbEntry, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Printf("user was not found: %v", err)
			return nil, err
		}

		log.Printf("unhandled error while searching for user record: %v", err)
		return nil, err
	}

	return &dbEntry.Name, nil
}

// GetPayment godoc
// @Summary Get a payment by ID
// @Param id path int true "Payment ID"
// @Router /api/v1/payment/{id} [get]
func (ctrl *Controller) GetPayment(c *gin.Context) {
	id := c.Param("id")

	var p models.PaymentDBEntry
	if err := ctrl.db.First(&p, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "payment not found"})
			return
		}

		log.Printf("db error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	c.JSON(http.StatusOK, p)
}

// GetPayments godoc
// @Summary Get all payments
// @Router /api/v1/payment/all [get]
func (ctrl *Controller) GetPayments(c *gin.Context) {
	var req []models.PaymentDBEntry

	if err := ctrl.db.Find(&req).Error; err != nil {
		log.Printf("db error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	c.JSON(http.StatusOK, req)
}

// PostPayment godoc
// @Summary      Create a new payment
// @Description  save a payment in the database
// @Accept       json
// @Produce      json
// @Param        payment  body      models.Payment  true  "payment data"
// @Success      200      {object}  models.Payment
// @Failure      400      {object}  map[string]string
// @Failure      500      {object}  map[string]string
// @Router       /api/v1/payment [post]
func (ctrl *Controller) PostPayment(c *gin.Context) {
	var req models.Payment
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// check if the payer exists and get their id
	var payer models.UserDBEntry
	if err := ctrl.db.Where("name = ?", req.Payer).First(&payer).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "payer not found"})
			return
		}

		log.Printf("db error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	// check if the owers exist and get their ids
	var owers []models.UserDBEntry
	for _, ower := range req.Owers {
		var owerEntry models.UserDBEntry

		// if ower == payer, then you've done something wrong
		if ower == payer.Name {
			c.JSON(http.StatusBadRequest, gin.H{"error": "payer cannot owe itself"})
		}

		if err := ctrl.db.Where("name = ?", ower).First(&owerEntry).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				c.JSON(http.StatusBadRequest, gin.H{"error": "ower not found"})
				return
			}
			log.Printf("db error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
			return
		}

		owers = append(owers, owerEntry)
	}

	// build the db req
	dbReq := models.PaymentDBEntry{
		Amount:      req.Amount,
		Description: req.Description,
		Date:        time.Now(),
		PayerID:     payer.ID,
		Owers:       owers,
	}

	// save the payment to the database
	if err := ctrl.db.Create(&dbReq).Error; err != nil {
		log.Printf("failed to create payment: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not save payment"})
		return
	}

	log.Printf("Received payment request: %+v\n", dbReq)
	c.JSON(http.StatusOK, req)
}

// DeletePayment godoc
// @Summary Delete a payment by ID
// @Param id path int true "Payment ID"
// @Router /api/v1/payment/{id} [delete]
func (ctrl *Controller) DeletePayment(c *gin.Context) {
	id := c.Param("id")

	var p models.PaymentDBEntry
	if err := ctrl.db.First(&p, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "payment not found"})
			return
		}

		log.Printf("db error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	if err := ctrl.db.Unscoped().Delete(&p).Error; err != nil {
		log.Printf("failed to delete payment: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not delete payment"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"info": "payment deleted"})
}

// GetDbHealth godoc
// @Summary Health check
// @Router /api/v1/health [get]
func (ctrl *Controller) GetDbHealth(c *gin.Context) {
	// Run simple query to verify connectivity
	var result int
	if err := ctrl.db.Raw("SELECT 1").Scan(&result).Error; err != nil {
		log.Printf("health check failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok", "value": result})
}

// PostUser godoc
// @Summary      Create a new user
// @Description  save a user in the database
// @Accept       json
// @Produce      json
// @Param        user  body      models.User  true  "user data"
// @Success      200      {object}  models.User
// @Failure      400      {object}  map[string]string
// @Failure      500      {object}  map[string]string
// @Router       /api/v1/user [post]
func (ctrl *Controller) PostUser(c *gin.Context) {
	var req models.User

	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	dbReq := models.UserDBEntry{
		Name: req.Name,
	}

	if err := ctrl.db.Create(&dbReq).Error; err != nil {
		log.Printf("failed to create user: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not create user"})
		return
	}

	log.Printf("Received user request: %+v\n", dbReq)
	c.JSON(http.StatusOK, req)
}

// DeleteUser godoc
// @Summary Delete a user by user by name
// @Param name path string true "user name"
// @Router /api/v1/user/{name} [delete]
func (ctrl *Controller) DeleteUser(c *gin.Context) {
	var payments []models.PaymentDBEntry

	name := c.Param("name")

	// if the user is a part of any payments, we should not delete them

	// First resolve the ID of the user
	userId, err := ctrl.resolveUserIdFromName(name)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}

		log.Printf("failed to find user to delete: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not find user to delete"})
		return
	}

	paymentQueryErr := ctrl.db.
		Preload("Owers").
		Where("payer_id = ?", *userId).
		Or("id IN (SELECT payment_db_entry_id FROM payment_owers WHERE user_db_entry_id = ?)", *userId).
		Find(&payments).Error
	if paymentQueryErr != nil {
		log.Printf("failed to find payments that the user is a part of: %v", paymentQueryErr)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not find payments for this user"})
		return
	}

	if len(payments) != 0 {
		c.JSON(http.StatusConflict, gin.H{"error": "cannot delete user because it is a part of active payments"})
		return
	}

	log.Printf("payments: %+v\n", payments)

	// delete the user
	if err := ctrl.db.Unscoped().Delete(&models.UserDBEntry{}, *userId).Error; err != nil {
		log.Printf("failed to delete user: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not delete user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"info": "all gone"})
}

// GetUser godoc
// @Summary Get a user by name
// @Param name path string true "User name"
// @Router /api/v1/user/{name} [get]
func (ctrl *Controller) GetUser(c *gin.Context) {
	name := c.Param("name")

	var p models.UserDBEntry
	if err := ctrl.db.Where("name = ?", name).First(&p).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}

		log.Printf("db error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error while fetching user"})
		return
	}

	c.JSON(http.StatusOK, p)
}

// GetUsers godoc
// @Summary Get all users
// @Router /api/v1/user/all [get]
func (ctrl *Controller) GetUsers(c *gin.Context) {
	var users []models.UserDBEntry

	if err := ctrl.db.Find(&users).Error; err != nil {
		log.Printf("db error while fetching users: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error while fetching users"})
		return
	}

	c.JSON(http.StatusOK, users)
}

// UpdateBalances godoc
// @Summary Update all balances
// @Router /api/v1/balance/all [post]
func (ctrl *Controller) UpdateBalances(c *gin.Context) {
	// to do this, we need to go through all the payments and calculate how much each user owes to each other user, then update the balances table accordingly
	type debtKey struct {
		From uint
		To   uint
	}

	var payments []models.PaymentDBEntry
	if err := ctrl.db.Preload("Owers").Find(&payments).Error; err != nil {
		log.Printf("db error while fetching payments: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error while fetching payments"})
		return
	}

	raw := map[debtKey]float64{}

	for _, paymentDbEntry := range payments {
		payerId := paymentDbEntry.PayerID
		owers := paymentDbEntry.Owers

		amountPerOwer := paymentDbEntry.Amount / float64(len(owers))

		for _, ower := range owers {
			raw[debtKey{From: ower.ID, To: payerId}] += amountPerOwer

			// Let's take this opportunity to also add an entry for the inverse if not already defined
			if _, ok := raw[debtKey{From: payerId, To: ower.ID}]; !ok {
				raw[debtKey{From: payerId, To: ower.ID}] = 0.0
			}
		}
	}

	// Now that we've gone through each payment and calculated the raw debts, we need to
	// calculate the ACTUAL debts. If A owes B $30 and B owes A $10, then you can rewrite
	// this as A owing B $20 and B owing A $0
	for debtEntry, amount := range raw {
		fromUser := debtEntry.From // person A
		toUser := debtEntry.To     // person B

		inverseDebt := raw[debtKey{From: toUser, To: fromUser}]

		if amount > inverseDebt {
			raw[debtEntry] = amount - inverseDebt
			raw[debtKey{From: toUser, To: fromUser}] = 0.0
		}
	}

	// When outputting to the API, we should use human-friendly names, not ids
	var response []models.Balance

	// Print the raw balances for debugging with the user names instead of ids
	for debtEntry, amount := range raw {
		if amount == 0 {
			continue // skip zero balances
		}

		fromUserName, err := ctrl.resolveUserNameFromId(debtEntry.From)
		if err != nil {
			log.Printf("db error while fetching user name from id for balance: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "db error while fetching user name from id for balance"})
			return
		}

		toUserName, err := ctrl.resolveUserNameFromId(debtEntry.To)
		if err != nil {
			log.Printf("db error while fetching user name from id for balance: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "db error while fetching user name from id for balance"})
			return
		}

		response = append(response, models.Balance{
			FromUser: *fromUserName,
			ToUser:   *toUserName,
			Amount:   amount,
		})

		log.Printf("Balance: %s owes %s: $%.2f\n", *fromUserName, *toUserName, amount)
	}

	c.JSON(http.StatusOK, response)
}

// GetBalances godoc
// @Summary Get all balances
// @Router /api/v1/balance/all [get]
func (ctrl *Controller) GetBalances(c *gin.Context) {
	// to do this, we need to go through all the payments and calculate how much each user owes to each other user, then update the balances table accordingly
	type debtKey struct {
		From uint
		To   uint
	}

	var payments []models.PaymentDBEntry
	if err := ctrl.db.Preload("Owers").Find(&payments).Error; err != nil {
		log.Printf("db error while fetching payments: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error while fetching payments"})
		return
	}

	raw := map[debtKey]float64{}

	for _, paymentDbEntry := range payments {
		payerId := paymentDbEntry.PayerID
		owers := paymentDbEntry.Owers

		amountPerOwer := paymentDbEntry.Amount / float64(len(owers))

		for _, ower := range owers {
			raw[debtKey{From: ower.ID, To: payerId}] += amountPerOwer

			// Let's take this opportunity to also add an entry for the inverse if not already defined
			if _, ok := raw[debtKey{From: payerId, To: ower.ID}]; !ok {
				raw[debtKey{From: payerId, To: ower.ID}] = 0.0
			}
		}
	}

	// Now that we've gone through each payment and calculated the raw debts, we need to
	// calculate the ACTUAL debts. If A owes B $30 and B owes A $10, then you can rewrite
	// this as A owing B $20 and B owing A $0
	for debtEntry, amount := range raw {
		fromUser := debtEntry.From // person A
		toUser := debtEntry.To     // person B

		inverseDebt := raw[debtKey{From: toUser, To: fromUser}]

		if amount > inverseDebt {
			raw[debtEntry] = amount - inverseDebt
			raw[debtKey{From: toUser, To: fromUser}] = 0.0
		}
	}

	// When outputting to the API, we should use human-friendly names, not ids
	var response []models.Balance

	// Print the raw balances for debugging with the user names instead of ids
	for debtEntry, amount := range raw {
		if amount == 0 {
			continue // skip zero balances
		}

		fromUserName, err := ctrl.resolveUserNameFromId(debtEntry.From)
		if err != nil {
			log.Printf("db error while fetching user name from id for balance: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "db error while fetching user name from id for balance"})
			return
		}

		toUserName, err := ctrl.resolveUserNameFromId(debtEntry.To)
		if err != nil {
			log.Printf("db error while fetching user name from id for balance: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "db error while fetching user name from id for balance"})
			return
		}

		response = append(response, models.Balance{
			FromUser: *fromUserName,
			ToUser:   *toUserName,
			Amount:   amount,
		})

		log.Printf("Balance: %s owes %s: $%.2f\n", *fromUserName, *toUserName, amount)
	}

	c.JSON(http.StatusOK, response)
}
