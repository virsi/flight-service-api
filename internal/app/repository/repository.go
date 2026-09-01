package repository

import "fmt"

type Repository struct {
}

func NewRepository() (*Repository, error) {
	return &Repository{}, nil
}

type Resource struct {
	ID          int
	Name        string  // наименование ресурса
	Description string  // краткое описание
	Status      string  // черновик / опубликован / удален
	ImageURL    string  // ссылка на изображение в Minio
	VideoURL    string  // ссылка на видео в Minio
	Unit        string  // единица измерения
	Price       float64 // цена за единицу
	Category    string  // категория ресурса
	CreatedAt   string  // дата создания
	Creator     string  // создатель
	FormedAt    string  // дата формирования
	Likes       []int   // вложенные лайки (ID пользователей)
}

func (r *Repository) GetResources() ([]Resource, error) {
	resources := []Resource{
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

	if len(resources) == 0 {
		return nil, fmt.Errorf("массив пустой")
	}

	return resources, nil
}

func (r *Repository) GetPublishedResources() ([]Resource, error) {
	resources, err := r.GetResources()
	if err != nil {
		return nil, err
	}

	var published []Resource
	for _, res := range resources {
		if res.Status == "опубликован" {
			published = append(published, res)
		}
	}

	return published, nil
}

func (r *Repository) GetResourcesByPrice(maxPrice float64) ([]Resource, error) {
	published, err := r.GetPublishedResources()
	if err != nil {
		return nil, err
	}

	var result []Resource
	for _, res := range published {
		if res.Price <= maxPrice {
			result = append(result, res)
		}
	}

	return result, nil
}

func (r *Repository) GetResource(id int) (Resource, error) {
	resources, err := r.GetResources()
	if err != nil {
		return Resource{}, err
	}

	for _, res := range resources {
		if res.ID == id {
			return res, nil
		}
	}

	return Resource{}, fmt.Errorf("ресурс не найден")
}

func (r *Repository) GetNextResourceID(id int) (int, error) {
	published, err := r.GetPublishedResources()
	if err != nil {
		return 0, err
	}

	if len(published) == 0 {
		return 0, fmt.Errorf("нет опубликованных ресурсов")
	}

	for _, res := range published {
		if res.ID > id {
			return res.ID, nil
		}
	}

	return published[0].ID, nil
}

func (r *Repository) GetDraft() (Resource, error) {
	resources, err := r.GetResources()
	if err != nil {
		return Resource{}, err
	}

	for _, res := range resources {
		if res.Status == "черновик" {
			return res, nil
		}
	}

	return Resource{}, fmt.Errorf("черновик не найден")
}
