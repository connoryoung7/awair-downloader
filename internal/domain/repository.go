package domain

type ReadingRepository interface {
	InsertReading(reading Reading) error
}
