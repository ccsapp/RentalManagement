package car

import (
	"RentalManagement/logic/model"
	carTypes "git.scc.kit.edu/cm-tm/cm-team/projectwork/pse/domain/d-cargotypes.git"
	openapiTypes "github.com/deepmap/oapi-codegen/pkg/types"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

var exampleDomainCar = carTypes.Car{
	Brand: "Volkswagen",
	DynamicData: carTypes.DynamicData{
		DoorsLockState:      carTypes.UNLOCKED,
		EngineState:         carTypes.OFF,
		FuelLevelPercentage: 23,
		Position: carTypes.DynamicDataPosition{
			Latitude:  49.0069,
			Longitude: 8.4037,
		},
		TrunkLockState: carTypes.UNLOCKED,
	},
	Model: "Golf",
	ProductionDate: openapiTypes.Date{
		Time: time.Date(2022, 12, 01, 0, 0, 0, 0, time.UTC),
	},
	TechnicalSpecification: carTypes.TechnicalSpecification{
		Color: "black",
		Consumption: carTypes.TechnicalSpecificationConsumption{
			City:     6.4,
			Combined: 5.2,
			Overland: 4.6,
		},
		Emissions: carTypes.TechnicalSpecificationEmissions{
			City:     120,
			Combined: 100,
			Overland: 90,
		},
		Engine: carTypes.TechnicalSpecificationEngine{

			Power: 110,
			Type:  "180 CDI",
		},
		Fuel:          carTypes.ELECTRIC,
		FuelCapacity:  "54.0L;85.2kWh",
		NumberOfDoors: 5,
		NumberOfSeats: 7,
		Tire: carTypes.TechnicalSpecificationTire{
			Manufacturer: "GOODYEAR",
			Type:         "185/65R15",
		},
		Transmission: carTypes.MANUAL,
		TrunkVolume:  435,
		Weight:       1320,
	},
	Vin: "3VW217AU9FM500158",
}

var exampleModelCar = model.Car{
	Brand: "Volkswagen",
	DynamicData: &model.DynamicData{
		DoorsLockState:      model.UNLOCKED,
		EngineState:         model.OFF,
		FuelLevelPercentage: 23,
		Position: carTypes.DynamicDataPosition{
			Latitude:  49.0069,
			Longitude: 8.4037,
		},
		TrunkLockState: model.UNLOCKED,
	},
	Model: "Golf",
	TechnicalSpecification: &model.TechnicalSpecification{
		Color: "black",
		Consumption: carTypes.TechnicalSpecificationConsumption{
			City:     6.4,
			Combined: 5.2,
			Overland: 4.6,
		},
		Emissions: carTypes.TechnicalSpecificationEmissions{
			City:     120,
			Combined: 100,
			Overland: 90,
		},
		Engine: carTypes.TechnicalSpecificationEngine{

			Power: 110,
			Type:  "180 CDI",
		},
		Fuel:          model.ELECTRIC,
		FuelCapacity:  "54.0L;85.2kWh",
		NumberOfDoors: 5,
		NumberOfSeats: 7,
		Transmission:  model.MANUAL,
		TrunkVolume:   435,
		Weight:        1320,
	},
	Vin: "3VW217AU9FM500158",
}

var exampleCarBase = model.Car{
	Brand:                  "Volkswagen",
	DynamicData:            nil,
	Model:                  "Golf",
	TechnicalSpecification: nil,
	Vin:                    "3VW217AU9FM500158",
}

var exampleCarAvailable = model.CarAvailable{
	Brand:         "Volkswagen",
	Model:         "Golf",
	NumberOfSeats: 7,
	Vin:           "3VW217AU9FM500158",
}

var exampleCarStatic = model.Car{
	Brand:       "Volkswagen",
	DynamicData: nil,
	Model:       "Golf",
	TechnicalSpecification: &model.TechnicalSpecification{
		Color: "black",
		Consumption: carTypes.TechnicalSpecificationConsumption{
			City:     6.4,
			Combined: 5.2,
			Overland: 4.6,
		},
		Emissions: carTypes.TechnicalSpecificationEmissions{
			City:     120,
			Combined: 100,
			Overland: 90,
		},
		Engine: carTypes.TechnicalSpecificationEngine{

			Power: 110,
			Type:  "180 CDI",
		},
		Fuel:          model.ELECTRIC,
		FuelCapacity:  "54.0L;85.2kWh",
		NumberOfDoors: 5,
		NumberOfSeats: 7,
		Transmission:  model.MANUAL,
		TrunkVolume:   435,
		Weight:        1320,
	},
	Vin: "3VW217AU9FM500158",
}

func TestMapToCar(t *testing.T) {
	assert.Equal(t, &exampleModelCar, MapToCar(&exampleDomainCar))
}

func TestMapToCarAvailable(t *testing.T) {
	assert.Equal(t, &exampleCarAvailable, MapToCarAvailable(&exampleDomainCar))
}

func TestMapToCarBase(t *testing.T) {
	assert.Equal(t, &exampleCarBase, MapToCarBase(&exampleDomainCar))
}

func TestMapToCarStatic(t *testing.T) {
	assert.Equal(t, &exampleCarStatic, MapToCarStatic(&exampleDomainCar))
}
