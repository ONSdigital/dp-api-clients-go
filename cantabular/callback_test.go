package cantabular

import (
	"net/http"
	"fmt"
	"errors"
	"testing"

	"github.com/ONSdigital/log.go/v2/log"

	. "github.com/smartystreets/goconvey/convey"
)

func TestCallbackHappy(t *testing.T) {

	Convey("Given an error with embedded status code", t, func() {
		err := &Error{
			statusCode: http.StatusBadRequest,
		}

		Convey("When StatusCode(err) is called", func() {
			statusCode := StatusCode(err)
			So(statusCode, ShouldEqual, http.StatusBadRequest)
		})
	})

	Convey("Given an error with embedded logData", t, func() {
		err := &Error{
			logData: log.Data{
				"log":"data",
			},
		}

		Convey("When LogData(err) is called", func() {
			logData := LogData(err)
			So(logData, ShouldResemble, log.Data{"log":"data"})
		})
	})

	Convey("Given an error chain with wrapped logData", t, func() {
		err1 := &Error{
			err: errors.New("original error"),
			logData: log.Data{
				"log":"data",
			},
		}

		err2 := &Error{
			err: fmt.Errorf("err1: %w", err1),
			logData: log.Data{
				"additional": "data",
			},
		}

		err3 := &Error{
			err: fmt.Errorf("err2: %w", err2),
			logData: log.Data{
				"final": "data",
			},
		}

		Convey("When UnwrapLogData(err) is called", func() {
			logData := UnwrapLogData(err3)
			expected := log.Data{
				"final":"data",
				"additional":"data",
				"log":"data",
			}

			So(logData, ShouldResemble,expected)
		})
	})

		Convey("Given an error chain with intermittent wrapped logData", t, func() {
		err1 := &Error{
			err: errors.New("original error"),
			logData: log.Data{
				"log":"data",
			},
		}

		err2 := &Error{
			err: fmt.Errorf("err1: %w", err1),
		}

		err3 := &Error{
			err: fmt.Errorf("err2: %w", err2),
			logData: log.Data{
				"final": "data",
			},
		}

		Convey("When UnwrapLogData(err) is called", func() {
			logData := UnwrapLogData(err3)
			expected := log.Data{
				"final":"data",
				"log":"data",
			}

			So(logData, ShouldResemble,expected)
		})
	})

	Convey("Given an error chain with wrapped logData with duplicate key values", t, func() {
		err1 := &Error{
			err: errors.New("original error"),
			logData: log.Data{
				"log":"data",
				"duplicate": "duplicate_data1",
				"request_id": "ADB45F",
			},
		}

		err2 := &Error{
			err: fmt.Errorf("err1: %w", err1),
			logData: log.Data{
				"additional": "data",
				"duplicate": "duplicate_data2",
				"request_id": "ADB45F",
			},
		}

		err3 := &Error{
			err: fmt.Errorf("err2: %w", err2),
			logData: log.Data{
				"final": "data",
				"duplicate": "duplicate_data3",
				"request_id": "ADB45F",
			},
		}

		Convey("When UnwrapLogData(err) is called", func() {
			logData := UnwrapLogData(err3)
			expected := log.Data{
				"final":"data",
				"additional":"data",
				"log":"data",
				"duplicate": []interface{}{
					"duplicate_data3",
					"duplicate_data2",
					"duplicate_data1",
				},
				"request_id": "ADB45F",
			}

			So(logData, ShouldResemble,expected)
		})
	})
}