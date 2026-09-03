package repository

import "fmt"

type Repository struct {
}

func NewRepository() (*Repository, error) {
	return &Repository{}, nil
}

type FlightService struct {
	ID          int
	Name        string
	Description string
	Status      string
	ImageURL    string
	VideoURL    string
	Unit        string
	Price       float64
	Category    string
	CreatedAt   string
	Creator     string
	FormedAt    string
	Likes       []int
}

func (r *Repository) GetFlightServices() ([]FlightService, error) {
	services := []FlightService{
		{
			ID: 1, Name: "Аэродромный тягач Goldhofer AST-2X",
			Description: "Буксировка и постановка воздушного судна на стоянку.",
			Status:      "опубликован",
			ImageURL:    "http://localhost:9100/flight-media/tug.jpg",
			VideoURL:    "http://localhost:9100/flight-media/tug.mp4",
			Unit:        "час", Price: 18500, Category: "Техника",
			CreatedAt: "2026-08-25 10:00", Creator: "диспетчер", FormedAt: "2026-08-25 10:05",
			Likes: []int{1, 2, 3, 4, 5},
		},
		{
			ID: 2, Name: "Авиатопливо ТС-1",
			Description: "Заправка воздушного судна авиационным керосином.",
			Status:      "опубликован",
			ImageURL:    "http://localhost:9100/flight-media/jet-fuel.jpg",
			VideoURL:    "http://localhost:9100/flight-media/jet-fuel.mp4",
			Unit:        "литр", Price: 82, Category: "Топливо",
			CreatedAt: "2026-08-25 10:00", Creator: "диспетчер", FormedAt: "2026-08-25 10:05",
			Likes: []int{1, 2, 3},
		},
		{
			ID: 3, Name: "Бортовое питание (эконом)",
			Description: "Комплект бортового питания на одного пассажира.",
			Status:      "опубликован",
			ImageURL:    "http://localhost:9100/flight-media/catering.jpg",
			VideoURL:    "http://localhost:9100/flight-media/catering.mp4",
			Unit:        "порция", Price: 460, Category: "Питание",
			CreatedAt: "2026-08-25 10:00", Creator: "диспетчер", FormedAt: "2026-08-25 10:05",
			Likes: []int{1, 2, 3, 4, 5, 6},
		},
		{
			ID: 4, Name: "Агент авиакомпании",
			Description: "Сопровождение рейса представителем авиакомпании.",
			Status:      "опубликован",
			ImageURL:    "http://localhost:9100/flight-media/airline-agent.jpg",
			VideoURL:    "http://localhost:9100/flight-media/airline-agent.mp4",
			Unit:        "час", Price: 2800, Category: "Персонал",
			CreatedAt: "2026-08-25 10:00", Creator: "диспетчер", FormedAt: "2026-08-25 10:05",
			Likes: []int{1, 2},
		},
		{
			ID: 5, Name: "Наземный источник питания (GPU)",
			Description: "Обеспечение самолёта электропитанием на стоянке.",
			Status:      "опубликован",
			ImageURL:    "http://localhost:9100/flight-media/gpu.jpg",
			VideoURL:    "http://localhost:9100/flight-media/gpu.mp4",
			Unit:        "час", Price: 9400, Category: "Техника",
			CreatedAt: "2026-08-25 10:00", Creator: "диспетчер", FormedAt: "2026-08-25 10:05",
			Likes: []int{1},
		},
		{
			ID: 6, Name: "Уборка салона",
			Description: "Полная уборка пассажирского салона перед рейсом.",
			Status:      "опубликован",
			ImageURL:    "http://localhost:9100/flight-media/cleaning.jpg",
			VideoURL:    "http://localhost:9100/flight-media/cleaning.mp4",
			Unit:        "рейс", Price: 5200, Category: "Персонал",
			CreatedAt: "2026-08-25 10:00", Creator: "диспетчер", FormedAt: "2026-08-25 10:05",
			Likes: []int{1, 2, 3, 4},
		},
		{
			ID: 7, Name: "Багажный тягач",
			Description: "Транспортировка багажа между терминалом и самолётом.",
			Status:      "черновик",
			ImageURL:    "http://localhost:9100/flight-media/baggage-tug.jpg",
			VideoURL:    "http://localhost:9100/flight-media/baggage-tug.mp4",
			Unit:        "час", Price: 7600, Category: "Техника",
			CreatedAt: "2026-08-25 10:00", Creator: "диспетчер", FormedAt: "",
			Likes: []int{},
		},
		{
			ID: 8, Name: "Противообледенительная обработка",
			Description: "Обработка воздушного судна противообледенительной жидкостью.",
			Status:      "удален",
			ImageURL:    "http://localhost:9100/flight-media/deicing.jpg",
			VideoURL:    "http://localhost:9100/flight-media/deicing.mp4",
			Unit:        "рейс", Price: 15000, Category: "Техника",
			CreatedAt: "2026-08-25 10:00", Creator: "диспетчер", FormedAt: "",
			Likes: []int{},
		},
	}

	if len(services) == 0 {
		return nil, fmt.Errorf("массив пустой")
	}

	return services, nil
}

func (r *Repository) GetPublishedFlightServices() ([]FlightService, error) {
	services, err := r.GetFlightServices()
	if err != nil {
		return nil, err
	}

	var published []FlightService
	for _, res := range services {
		if res.Status == "опубликован" {
			published = append(published, res)
		}
	}

	return published, nil
}

func (r *Repository) GetFlightServicesByPrice(maxPrice float64) ([]FlightService, error) {
	published, err := r.GetPublishedFlightServices()
	if err != nil {
		return nil, err
	}

	var result []FlightService
	for _, res := range published {
		if res.Price <= maxPrice {
			result = append(result, res)
		}
	}

	return result, nil
}

func (r *Repository) GetFlightService(id int) (FlightService, error) {
	services, err := r.GetFlightServices()
	if err != nil {
		return FlightService{}, err
	}

	for _, res := range services {
		if res.ID == id {
			return res, nil
		}
	}

	return FlightService{}, fmt.Errorf("услуга не найдена")
}

func (r *Repository) GetNextFlightServiceID(id int) (int, error) {
	published, err := r.GetPublishedFlightServices()
	if err != nil {
		return 0, err
	}

	if len(published) == 0 {
		return 0, fmt.Errorf("нет опубликованных услуг")
	}

	for _, res := range published {
		if res.ID > id {
			return res.ID, nil
		}
	}

	return published[0].ID, nil
}

func (r *Repository) GetDraftFlightService() (FlightService, error) {
	services, err := r.GetFlightServices()
	if err != nil {
		return FlightService{}, err
	}

	for _, res := range services {
		if res.Status == "черновик" {
			return res, nil
		}
	}

	return FlightService{}, fmt.Errorf("черновик не найден")
}
