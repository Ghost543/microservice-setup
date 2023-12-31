package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"log"
	"net/http"
	"time"
)

type OrderStatus int8

const (
	Pending OrderStatus = iota
	Booked
	Taken
	Cancelled
	Returned
)

func (o OrderStatus) String() string {
	switch o {
	case Pending:
		return "pending"
	case Booked:
		return "booked"
	case Taken:
		return "taken"
	case Cancelled:
		return "cancelled"
	case Returned:
		return "returned"
	}
	return "unknown"
}

type Product struct {
	Id       uint    `json:"id"`
	Name     string  `json:"name"`
	Cost     float64 `json:"cost"`
	Quantity uint8   `json:"quantity"`
	country  string
	city     string
}

type Customer struct {
	Name    string `json:"name"`
	Age     uint16 `json:"age"`
	Tel     string `json:"tel"`
	Email   string `json:"email"`
	Country string `json:"country"`
	City    string `json:"city"`
	Address string `json:"address"`
}

type Order struct {
	Id             uint        `json:"id"`
	CustomerId     uint        `json:"customer_id"`
	Status         OrderStatus `json:"status" default:"0"`
	ShoppingCartId uint        `json:"shopping_cart_id"`
	TotalCost      float64     `json:"totalCost"`
	Customer       Customer    `json:"customer" gorm:"-" default:"{}"`
	Products       []Product   `json:"products " gorm:"-" default:"[]"`
}

func main() {
	db, err_ := gorm.Open(postgres.Open("host=127.0.0.1 user=postgres password=root dbname=postgres port=5432 sslmode=disable TimeZone=Asia/Shanghai"), &gorm.Config{})

	if err_ != nil {
		panic("Failed to connect to DB")
	}

	err := db.AutoMigrate(&Order{})
	if err != nil {
		return
	}

	app := fiber.New()
	app.Use(cors.New())
	app.Use(compress.New())
	app.Use(logger.New())

	app.Get("api/orders", func(ctx *fiber.Ctx) error {
		var orders []Order
		db.Find(&orders)
		for i, order := range orders {
			res, err := http.Get(fmt.Sprintf("http://127.0.0.1:8082/api/customers/%d", order.CustomerId))
			if err != nil {
				return err
			}
			resP, err__ := http.Get(fmt.Sprintf("http://127.0.0.1:8082/api/shopping_cart/%d", order.ShoppingCartId))
			if err__ != nil {
				return err__
			}

			var products []Product
			er := json.NewDecoder(resP.Body).Decode(&products)
			var customer Customer
			err_ := json.NewDecoder(res.Body).Decode(&customer)
			if err_ != nil {
				return err_
			}
			if er != nil {
				return er
			}
			orders[i].Products = products
			orders[i].Customer = customer
		}

		return ctx.JSON(orders)
	})

	app.Get("api/orders/:id", func(ctx *fiber.Ctx) error {
		var order Order
		db.First(&order, "id = ?", ctx.Params("id"))

		res, err := http.Get(fmt.Sprintf("http://127.0.0.1:8082/api/customers/%d", order.CustomerId))
		if err != nil {
			return err
		}
		resP, err__ := http.Get(fmt.Sprintf("http://127.0.0.1:8082/api/shopping_cart/%d", order.ShoppingCartId))
		if err__ != nil {
			return err__
		}

		var products []Product
		er := json.NewDecoder(resP.Body).Decode(&products)
		var customer Customer
		err_ := json.NewDecoder(res.Body).Decode(&customer)
		if err_ != nil {
			return err_
		}
		if er != nil {
			return er
		}
		order.Products = products
		order.Customer = customer

		return ctx.Status(200).JSON(&fiber.Map{
			"status": "Fetch order",
			"order":  order,
		})
	})

	app.Post("orders/request", func(ctx *fiber.Ctx) error {
		var order Order
		if err := ctx.BodyParser(&order); err != nil {
			return err
		}
		db.Create(&order)
		return ctx.Status(201).JSON(&fiber.Map{
			"status": "Created",
			"order":  order,
		})
	})

	app.Get("api/orders/:id/destination", func(ctx *fiber.Ctx) error {
		var order Order
		db.First(&order, "id = ?", ctx.Params("id"))
		res, err := http.Get(fmt.Sprintf("http://127.0.0.1:8082/api/customers/%d", order.CustomerId))
		if err != nil {
			return err
		}
		resP, err__ := http.Get(fmt.Sprintf("http://127.0.0.1:8082/api/shopping_cart/%d", order.ShoppingCartId))
		if err__ != nil {
			return err__
		}
		var products []Product
		er := json.NewDecoder(resP.Body).Decode(&products)
		var customer Customer
		err_ := json.NewDecoder(res.Body).Decode(&customer)
		if err_ != nil {
			return err_
		}
		if er != nil {
			return er
		}
		order.Products = products
		order.Customer = customer
		return ctx.Status(200).JSON(&fiber.Map{
			"Destination": fmt.Sprintf("Country: %s, City: %s, Address: %s", order.Customer.Country, order.Customer.City, order.Customer.Address),
		})
	})

	app.Patch("api/orders/:id", func(ctx *fiber.Ctx) error {
		var order Order
		var updates Order

		db.First(&order, "id = ?", ctx.Params("id"))

		if err := ctx.BodyParser(&updates); err != nil {
			return err
		}
		order.ShoppingCartId = updates.ShoppingCartId
		order.Customer = updates.Customer
		order.Status = updates.Status
		order.TotalCost = updates.TotalCost

		db.Save(&order)

		return ctx.Status(200).JSON(&fiber.Map{
			"status": "Update",
			"order":  order,
		})
	})

	app.Delete("api/orders/:id", func(ctx *fiber.Ctx) error {
		db.Delete(&Order{}, "id = ?", ctx.Params("id"))
		return ctx.Status(200).JSON(&fiber.Map{
			"status": "Deleted",
		})
	})

	app.Get("api/orders/:id/shipping", func(ctx *fiber.Ctx) error {
		var order Order
		db.First(&order, "id = ?", ctx.Params("id"))
		var req map[string]uint
		req = make(map[string]uint)
		req["order_id"] = order.Id
		req["customer_id"] = order.CustomerId
		buff, err := json.Marshal(req)
		if err != nil {
			panic(err)
		}
		request, err := http.NewRequest("POST", "https://localhost:8080/api/shipping/receive", bytes.NewReader(buff))
		if err != nil {
			return err
		}
		request.Header.Set("Content-Type", "application/json")
		client := http.Client{Timeout: 10 * time.Second}
		do, err := client.Do(request)
		if err != nil {
			return err
		}

		return ctx.JSON(do)
	})
	app.Get("api/orders/:id/notify", func(ctx *fiber.Ctx) error {
		var order Order
		db.First(&order, "id = ?", ctx.Params("id"))
		req := map[string]interface{}{
			"customer_id": order.CustomerId,
			"message":     fmt.Sprintf("order %s", order.Status.String()),
		}

		buff, err := json.Marshal(req)
		if err != nil {
			panic(err)
		}
		request, err := http.NewRequest("POST", "https://localhost:8081/notification/send", bytes.NewReader(buff))
		if err != nil {
			return err
		}
		request.Header.Set("Content-Type", "application/json")
		client := http.Client{Timeout: 10 * time.Second}
		do, err := client.Do(request)
		if err != nil {
			return err
		}

		return ctx.JSON(do)
	})

	log.Fatal(app.Listen(":8001"))

}
