package operations

import (
	"RentalManagement/infrastructure/car"
	"RentalManagement/infrastructure/database"
)

type operations struct {
	carClient car.ClientWithResponsesInterface
	crud      database.ICRUD
}

func NewOperations(carClient car.ClientWithResponsesInterface, crud database.ICRUD) IOperations {
	return &operations{
		carClient: carClient,
		crud:      crud,
	}
}
