package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	_ "github.com/go-sql-driver/mysql"
)

// Event representa la estructura de un evento
type Event struct {
	ID          int     `json:"id"`
	Name        string  `json:"name"`
	Price       float64 `json:"price"`
	Description string  `json:"description"`
	EventType   string  `json:"event_type"`
	Date        string  `json:"date"`
}

var events []Event

func main() {
	// Conectar a la base de datos MySQL
	db, err := sql.Open("mysql", "root:@tcp(127.0.0.1:3306)/basededatos")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Crear una instancia de Fiber
	app := fiber.New()

	// Middleware de registro
	app.Use(logger.New())

	// Rutas RESTful para eventos
	app.Get("/eventos", func(c *fiber.Ctx) error {
		verEventos(db)
		return c.JSON(events)
	})

	app.Post("/eventos", func(c *fiber.Ctx) error {
		agregarEvento(db, c)
		return c.SendString("Evento creado exitosamente.")
	})

	app.Put("/eventos/:id", func(c *fiber.Ctx) error {
		id := c.Params("id")
		editarEvento(db, id, c)
		return c.SendString("Evento actualizado exitosamente.")
	})

	app.Delete("/eventos/:id", func(c *fiber.Ctx) error {
		id := c.Params("id")
		eliminarEvento(db, id, c)
		return c.SendString("Evento eliminado exitosamente.")
	})

	for {
		fmt.Println("Selecciona la opción deseada:")
		fmt.Println("1. Ver evento")
		fmt.Println("2. Agregar evento")
		fmt.Println("3. Editar evento")
		fmt.Println("4. Eliminar evento")
		fmt.Println("5. Salir")

		var opcion int
		fmt.Print("Opción: ")
		_, err := fmt.Scan(&opcion)
		if err != nil {
			fmt.Println("Error al leer la opción:", err)
			continue
		}

		switch opcion {
		case 1:
			fmt.Println("Seleccionaste 'Ver evento'")
			verEventos(db)
		case 2:
			fmt.Println("Seleccionaste 'Agregar evento'")
			agregarEvento(db, nil)
		case 3:
			fmt.Println("Seleccionaste 'Editar evento'")
			editarEvento(db, "", nil)
		case 4:
			fmt.Println("Seleccionaste 'Eliminar evento'")
			eliminarEvento(db, "", nil)
		case 5:
			fmt.Println("Saliendo...")
			os.Exit(0)
		default:
			fmt.Println("Opción no válida. Por favor, selecciona una opción válida.")
		}
	}

	app.Listen(":3000")
}

func verEventos(db *sql.DB) {
	rows, err := db.Query("SELECT * FROM eventos")
	if err != nil {
		fmt.Println("Error al obtener los eventos:", err)
		return
	}
	defer rows.Close()

	events = make([]Event, 0)
	for rows.Next() {
		event := Event{}
		if err := rows.Scan(&event.ID, &event.Name, &event.Price, &event.Description, &event.EventType, &event.Date); err != nil {
			fmt.Println("Error al escanear los eventos:", err)
			return
		}
		events = append(events, event)
	}

	for _, event := range events {
		fmt.Printf("ID: %d, Nombre: %s, Precio: %.2f, Descripción: %s, Tipo: %s, Fecha: %s\n", event.ID, event.Name, event.Price, event.Description, event.EventType, event.Date)
	}
}

func agregarEvento(db *sql.DB, c *fiber.Ctx) {
	event := new(Event)
	if c != nil {
		if err := c.BodyParser(event); err != nil {
			c.Status(fiber.StatusBadRequest).SendString("Error al parsear la solicitud")
			return
		}
	} else {
		fmt.Print("Nombre del evento: ")
		fmt.Scan(&event.Name)
		fmt.Print("Precio: ")
		fmt.Scan(&event.Price)
		fmt.Print("Descripción: ")
		fmt.Scan(&event.Description)
		fmt.Print("Tipo de evento: ")
		fmt.Scan(&event.EventType)
		fmt.Print("Fecha: ")
		fmt.Scan(&event.Date)
	}

	_, err := db.Exec("INSERT INTO eventos (name, price, description, event_type, date) VALUES (?, ?, ?, ?, ?)",
		event.Name, event.Price, event.Description, event.EventType, event.Date)
	if err != nil {
		if c != nil {
			c.Status(fiber.StatusInternalServerError).SendString("Error al crear el evento")
		} else {
			fmt.Println("Error al crear el evento:", err)
		}
		return
	}

	events = append(events, *event)

	if c != nil {
		c.Status(fiber.StatusCreated).SendString("Evento creado exitosamente.")
	} else {
		fmt.Println("Evento creado exitosamente.")
	}
}

func editarEvento(db *sql.DB, id string, c *fiber.Ctx) {
	event := new(Event)
	if c != nil {
		if err := c.BodyParser(event); err != nil {
			c.Status(fiber.StatusBadRequest).SendString("Error al parsear la solicitud")
			return
		}
	} else {
		fmt.Print("Nuevo nombre del evento: ")
		fmt.Scan(&event.Name)
		fmt.Print("Nuevo precio: ")
		fmt.Scan(&event.Price)
		fmt.Print("Nueva descripción: ")
		fmt.Scan(&event.Description)
		fmt.Print("Nuevo tipo de evento: ")
		fmt.Scan(&event.EventType)
		fmt.Print("Nueva fecha: ")
		fmt.Scan(&event.Date)
	}

	_, err := db.Exec("UPDATE eventos SET name=?, price=?, description=?, event_type=?, date=? WHERE id=?",
		event.Name, event.Price, event.Description, event.EventType, event.Date, id)
	if err != nil {
		if c != nil {
			c.Status(fiber.StatusInternalServerError).SendString("Error al actualizar el evento")
		} else {
			fmt.Println("Error al actualizar el evento:", err)
		}
		return
	}

	for i, e := range events {
		if e.ID == event.ID {
			events[i] = *event
			break
		}
	}

	if c != nil {
		c.SendString("Evento actualizado exitosamente.")
	} else {
		fmt.Println("Evento actualizado exitosamente.")
	}
}

func eliminarEvento(db *sql.DB, id string, c *fiber.Ctx) {
	if id == "" && c != nil {
		c.Status(fiber.StatusBadRequest).SendString("ID del evento no especificado")
		return
	}

	if id == "" {
		fmt.Print("ID del evento a eliminar: ")
		fmt.Scan(&id)
	}

	_, err := db.Exec("DELETE FROM eventos WHERE id=?", id)
	if err != nil {
		if c != nil {
			c.Status(fiber.StatusInternalServerError).SendString("Error al eliminar el evento")
		} else {
			fmt.Println("Error al eliminar el evento:", err)
		}
		return
	}

	if c != nil {
		c.SendString("Evento eliminado exitosamente.")
	} else {
		fmt.Println("Evento eliminado exitosamente.")
	}
}
