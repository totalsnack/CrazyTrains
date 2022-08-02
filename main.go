package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"sort"
	"time"
)

type Trains []Train

type Train struct {
	TrainID            int       `json:"trainId"`
	DepartureStationID int       `json:"departureStationId"`
	ArrivalStationID   int       `json:"arrivalStationId"`
	Price              float32   `json:"price"`
	ArrivalTime        time.Time `json:"arrivalTime"`
	DepartureTime      time.Time `json:"departureTime"`
}

type Config struct {
	ArrivalStationID   string `json:"arrivalStationId"`
	DepartureStationID string `json:"departureStationId"`
	Criteria           string `json:"criteria"`
}

const (
	pathToData     = "./data.json"
	pathToUserData = "./config.json"
)

// Sorting criteria
const (
	priceCriteria     = "price"          // спершу дешеві
	arrivalCriteria   = "arrival-time"   // спершу ті, що раніше прибувають
	departureCriteria = "departure-time" // спершу ті, що раніше відправляються
)

const maxNumOfTrains = 3

// HHMMSS 24h Missing time format
const HHMMSS = "15:04:05"

// These errors may be returned by entering user data
var (
	ErrCriteria       = errors.New("unsupported criteria")
	ErrEmptyDeparture = errors.New("empty departure station")
	ErrEmptyArrival   = errors.New("empty arrival station")
	ErrBadArrival     = errors.New("bad arrival station input")
	ErrBadDeparture   = errors.New("bad departure station input")
)

func main() {
	//	... запит даних від користувача
	config := &Config{}
	if err := parseJson(pathToUserData, config); err != nil {
		log.Fatal(err)
	}
	departureStation, arrivalStation, criteria := config.DepartureStationID, config.ArrivalStationID, config.Criteria
	result, err := FindTrains(departureStation, arrivalStation, criteria)

	//	... обробка помилки
	if err != nil {
		log.Fatal(err)
	}

	//	... друк result
	fmt.Println(result)
}

func FindTrains(departureStation, arrivalStation, criteria string) (Trains, error) {
	var needful, trains Trains

	// Input data
	err := parseJson(pathToData, &trains)
	if err != nil {
		err = fmt.Errorf("error parsing data: %w", err)
		return nil, err
	}

	// Sort trains by criteria
	err = trains.SortByCriteria(criteria)
	if err != nil {
		return nil, err
	}

	// Input data types conversion
	a, err := StrToN(departureStation)
	if err != nil {
		if len(departureStation) == 0 {
			return nil, ErrEmptyDeparture
		}
		return nil, ErrBadDeparture
	}

	b, err := StrToN(arrivalStation)
	if err != nil {
		if len(arrivalStation) == 0 {
			return nil, ErrEmptyArrival
		}
		return nil, ErrBadArrival
	}

	// Filter needful trains
	for _, t := range trains {
		if t.DepartureStationID == a && t.ArrivalStationID == b {
			needful = append(needful, t)
			if len(needful) == maxNumOfTrains {
				break
			}
		}
	}

	return needful, nil // маєте повернути правильні значення // i hope so
}

func (ts Trains) String() string {
	str, err := json.MarshalIndent(ts, "", " ")
	if err != nil {
		return fmt.Sprintf("%#v\n", ts)
	}
	return fmt.Sprintf("%s\n", str)
}

func (ts Trains) SortByCriteria(criteria string) (err error) {
	sortFilter := func(i int, j int) bool {
		switch criteria {
		case priceCriteria:
			return ts[i].Price < ts[j].Price
		case arrivalCriteria:
			return ts[i].ArrivalTime.Before(ts[j].ArrivalTime)
		case departureCriteria:
			return ts[i].DepartureTime.Before(ts[j].DepartureTime)
		default:
			err = ErrCriteria
		}
		return false
	}
	sort.SliceStable(ts, sortFilter)
	return err
}

// StrToN StrToInt converts  string to natural number
func StrToN(str string) (int, error) {
	var (
		num rune
		err = errors.New("bad number")
	)
	for _, r := range str {
		if r >= '0' && r <= '9' {
			num *= 10
			num += r - '0'
		} else {
			return 0, err
		}
	}
	if num != 0 {
		return int(num), nil
	}
	return 0, err
}
func parseTime(str string) (time.Time, error) {
	tm, err := time.Parse(HHMMSS, str)
	if err != nil {
		return tm, fmt.Errorf("error parsing string %s to time from json: %w", str, err)
	}
	return tm, nil
}

func parseJson(path string, i any) error {
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("error opening config file: %w", err)
	}

	defer func() {
		if err = file.Close(); err != nil {
			log.Printf("error closing config file: %v", err)
		}
	}()

	err = json.NewDecoder(file).Decode(&i)
	if err != nil {
		return fmt.Errorf("error decoding json from file: %w", err)
	}
	return err
}

// UnmarshalJSON unmarshal data to json aware struct
func (t *Train) UnmarshalJSON(data []byte) (err error) {
	// It's lame, but we do not want play with reflection, right?
	type clone Train
	raw := &struct {
		*clone
		ArrivalTime   string `json:"arrivalTime"`
		DepartureTime string `json:"departureTime"`
	}{
		clone: (*clone)(t),
	}
	if err = json.Unmarshal(data, &raw); err != nil {
		return err
	}

	if t.ArrivalTime, err = parseTime(raw.ArrivalTime); err != nil {
		return err
	}
	if t.DepartureTime, err = parseTime(raw.DepartureTime); err != nil {
		return err
	}
	return nil
}
