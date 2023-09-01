package main

import (
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Task struct {
	ID          uint           `gorm:"primaryKey"`
	Title       string         `gorm:"type:VARCHAR(255)"`
	Status      string         `gorm:"type:VARCHAR(10); DEFAULT:'ongoing'; CHECK:status IN ('ongoing', 'completed')"`
	Deadline    *time.Time     
	CreatedAt   time.Time      `gorm:"default:CURRENT_TIMESTAMP"`
	CompletedAt *time.Time     
	Subtasks    []Subtask 
}

type Subtask struct {
	ID          uint           `gorm:"primaryKey"`
	TaskID      uint           `gorm:"index"`
	Title       string         `gorm:"type:VARCHAR(255)"`
	Status      string         `gorm:"type:VARCHAR(10); DEFAULT:'ongoing'; CHECK:status IN ('ongoing', 'completed')"`
	CompletedAt *time.Time     
}

func main() {
	// Database config
	dsn := "host=localhost user=postgres password=postgre dbname=todoapp port=5432 sslmode=disable TimeZone=UTC"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		panic("Failed to connect to database")
	}

	// Automigrate
	db.AutoMigrate(&Task{}, &Subtask{})

	// API
	app := fiber.New()
	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Hello, World 👋!")
	})

	// Add Task
	app.Post("/tasks", func(c *fiber.Ctx) error {
		var task Task
		err := c.BodyParser(&task)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"message": "Invalid request body",
			})
		}
		err = db.Create(&task).Error
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"message": "Failed to create task",
			})
		}
		return c.Status(fiber.StatusCreated).JSON(fiber.Map{
			"message": "Task created successfully",
			"task": task,
		})
	})

	// Show Ongoing Task
	app.Get("/tasks/ongoing", func(c *fiber.Ctx) error {
		var task []Task
		err := db.Where("status = ?", "ongoing").Preload("Subtasks").Find(&task).Error
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"message": "Failed to get data",
			})
		}
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"task": task,
		})
	})

	// Delete Task
	app.Delete("tasks/:id", func(c *fiber.Ctx) error {
		var subtask []Subtask
		err := db.Where("task_id = ?", c.Params("id")).Find(&subtask).Error
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"message": "Failed to get data",
			})
		}
		for _, subtask := range subtask {
			err = db.Delete(&subtask).Error
			if err != nil {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"message": "Failed to delete subtask",
				})
			}
		}
		var task Task
		err = db.Delete(&task, c.Params("id")).Error
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"message": "Failed to delete data",
			})
		}
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"message": "Task deleted successfully",
		})
	})

	// Edit task
	app.Patch("tasks/:id", func(c *fiber.Ctx) error {
		var requestBody struct  {
			Title string `json:"title"`
		} 
		err := c.BodyParser(&requestBody)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"message": "Invalid request body",
			})
		}
		var task Task
		err = db.First(&task, c.Params("id")).Error
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"message": "Failed to get data",
			})
		}
		task.Title = requestBody.Title
		err = db.Save(&task).Error
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"message": "Failed to save data",
			})
		}
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"message": "Task successfully updated",
			"task": task,
		})
	})

	//Completing Task
	app.Patch("tasks/done/:id", func(c *fiber.Ctx) error {
		var task Task
		err = db.First(&task, c.Params("id")).Error
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"message": "Failed to get data",
			})
		}
		task.Status = "completed"
		completedAt := time.Now()
		task.CompletedAt = &completedAt
		
		err = db.Save(&task).Error
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"message": "Failed to save data",
			})
		}
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"message": "Task successfully completed",
			"task": task,
		})
	})

	// Show Completed Task
	app.Get("/tasks/completed", func(c *fiber.Ctx) error {
		var task []Task
		err := db.Where("status = ?", "completed").Preload("Subtasks").Find(&task).Error
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"message": "Failed to get data",
			})
		}
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"task": task,
		})
	})

	// Add Date and Time for Task
	app.Patch("tasks/deadline/:id", func(c *fiber.Ctx) error {
		var requestBody struct  {
			Time string `json:"time"`
		} 
		err := c.BodyParser(&requestBody)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"message": "Invalid request body",
			})
		}
		var task Task
		err = db.First(&task, c.Params("id")).Error
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"message": "Failed to get data",
			})
		}
		deadline, err := time.Parse("2006-01-02", requestBody.Time)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"message": "Invalid date format",
			})
		}
		task.Deadline = &deadline
		err = db.Save(&task).Error
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"message": "Failed to save data",
			})
		}
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"message": "Task successfully updated",
			"task": task,
		})
	})

	// Add Subtask
	app.Post("/subtasks/:idTask", func(c *fiber.Ctx) error {
		var subtask Subtask
		
		err := c.BodyParser(&subtask)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"message": "Invalid request body",
			})
		}
		idTask := c.Params("idTask")
		idTaskInt, err := strconv.Atoi(idTask)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"message": "Invalid Task ID",
			})
		}
		subtask.TaskID = uint(idTaskInt)
		err = db.Create(&subtask).Error
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"message": "Failed to create task",
			})
		}
		return c.Status(fiber.StatusCreated).JSON(fiber.Map{
			"message": "Subtask created successfully",
			"subtask": subtask,
		})
	})

	// Delete Subtask
	app.Delete("subtasks/:id", func(c *fiber.Ctx) error {
		var subtask Subtask
		err = db.Delete(&subtask, c.Params("id")).Error
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"message": "Failed to delete subtask",
			})
		}
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"message": "Subtask deleted successfully",
		})
	})

	// Edit Subtask
	app.Patch("subtasks/:id", func(c *fiber.Ctx) error {
		var requestBody struct  {
			Title string `json:"title"`
		} 
		err := c.BodyParser(&requestBody)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"message": "Invalid request body",
			})
		}
		var subtask Subtask
		err = db.First(&subtask, c.Params("id")).Error
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"message": "Failed to get subtask",
			})
		}
		subtask.Title = requestBody.Title
		err = db.Save(&subtask).Error
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"message": "Failed to save data",
			})
		}
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"message": "Subtask successfully updated",
			"task": subtask,
		})
	})

	//Completing Subtask
	app.Patch("subtasks/done/:id", func(c *fiber.Ctx) error {
		var subtask Subtask
		err = db.First(&subtask, c.Params("id")).Error
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"message": "Failed to get data",
			})
		}
		subtask.Status = "completed"
		completedAt := time.Now()
		subtask.CompletedAt = &completedAt
		err = db.Save(&subtask).Error
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"message": "Failed to save subtask",
			})
		}
		// Check all subtask completed
		var task Task
		err = db.Preload("Subtasks").First(&task, subtask.TaskID).Error
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"message": "Failed to get parent task",
			})
		}
		var flaggingCompleted = true
		for _, st := range task.Subtasks {
			if st.Status != "completed" {
				flaggingCompleted = false
				break
			}
		}
		if flaggingCompleted {
			task.Status = "completed"
			completedAt := time.Now()
			task.CompletedAt = &completedAt
			err = db.Save(&task).Error
			if err != nil {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"message": "Failed to update parent task",
				})
			}
		}
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"message": "Subtask successfully completed",
			"task": subtask,
		})
	})
	
	app.Listen(":3000")
}