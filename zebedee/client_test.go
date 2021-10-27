package zebedee

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/ONSdigital/dp-api-clients-go/v2/health"
	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	"github.com/ONSdigital/dp-mocking/httpmocks"
	dphttp "github.com/ONSdigital/dp-net/http"
	dprequest "github.com/ONSdigital/dp-net/request"
	"github.com/ONSdigital/log.go/v2/log"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	. "github.com/smartystreets/goconvey/convey"
)

const testHost = "http://localhost:8080"

var (
	testAccessToken = "test-access-token"
	initialState    = health.CreateCheckState(service)
)

func mockZebedeeServer(port chan int) {
	r := mux.NewRouter()

	r.Path("/data").HandlerFunc(contentData)
	r.Path("/parents").HandlerFunc(parents)
	r.Path("/filesize").HandlerFunc(filesize)

	l, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		log.Fatal(context.Background(), "error listening on local network address", err)
		os.Exit(2)
	}

	port <- l.Addr().(*net.TCPAddr).Port
	close(port)

	if err := http.Serve(l, r); err != nil {
		log.Fatal(context.Background(), "error serving http connections", err)
		os.Exit(2)
	}
}

func newMockHTTPClient(r *http.Response, err error) *dphttp.ClienterMock {
	return &dphttp.ClienterMock{
		SetPathsWithNoRetriesFunc: func(paths []string) {
			return
		},
		DoFunc: func(ctx context.Context, req *http.Request) (*http.Response, error) {
			return r, err
		},
		GetPathsWithNoRetriesFunc: func() []string {
			return []string{"/healthcheck"}
		},
	}
}

func newZebedeeClient(httpClient *dphttp.ClienterMock) *Client {
	healthClient := health.NewClientWithClienter("", testHost, httpClient)
	zebedeeClient := NewWithHealthClient(healthClient)
	return zebedeeClient
}

func contentData(w http.ResponseWriter, req *http.Request) {
	uri := req.URL.Query().Get("uri")

	serviceAuthToken := req.Header.Get(dprequest.FlorenceHeaderKey)
	if serviceAuthToken != testAccessToken {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("401 - No access token header set!"))
	}

	switch uri {
	case "foo":
		w.Write([]byte(`{}`))
	case "labour":
		w.Write([]byte(`{"downloads":[{"title":"Latest","file":"/employmentandlabourmarket/peopleinwork/workplacedisputesandworkingconditions/datasets/labourdisputesbysectorlabd02/labd02jul2015_tcm77-408195.xls"}],"section":{"markdown":""},"relatedDatasets":[{"uri":"/employmentandlabourmarket/peopleinwork/workplacedisputesandworkingconditions/datasets/labourdisputeslabd01"},{"uri":"/employmentandlabourmarket/peopleinwork/workplacedisputesandworkingconditions/datasets/stoppagesofworklabd03"}],"relatedDocuments":[{"uri":"/employmentandlabourmarket/peopleinwork/employmentandemployeetypes/bulletins/uklabourmarket/2015-07-15"}],"relatedMethodology":[],"type":"dataset_landing_page","uri":"/employmentandlabourmarket/peopleinwork/workplacedisputesandworkingconditions/datasets/labourdisputesbysectorlabd02","description":{"title":"Labour disputes by sector: LABD02","summary":"Labour disputes by sector.","keywords":["strike"],"metaDescription":"Labour disputes by sector.","nationalStatistic":true,"contact":{"email":"richard.clegg@ons.gsi.gov.uk\n","name":"Richard Clegg\n","telephone":"+44 (0)1633 455400 \n"},"releaseDate":"2015-07-14T23:00:00.000Z","nextRelease":"12 August 2015","datasetId":"","unit":"","preUnit":"","source":""}}`))
	case "12345":
		w.Write([]byte(`{"type":"dataset","uri":"www.google.com","downloads":[{"file":"test.txt"}],"supplementaryFiles":[{"title":"helloworld","file":"helloworld.txt"}],"versions":[{"uri":"www.google.com"}]}`))
	case "absoluteDownloadURI":
		w.Write([]byte(`{"type":"dataset","uri":"localhost","downloads":[{"file":"absoluteDownloadURI/test.txt"}],"supplementaryFiles":[{"title":"helloworld","file":"helloworld.txt"}],"versions":[{"uri":"www.google.com"}]}`))
	case "pageTitle":
		w.Write([]byte(`{"title":"baby-names","edition":"2017"}`))
	case "bulletin":
		w.Write([]byte(`{"relatedBulletins":[{"uri":"/economy/environmentalaccounts/datasets/ukenvironmentalaccountsenergybridging"}],"sections":[{"title":"Main points","markdown":"Main points markdown"},{"title":"Overview","markdown":"Overview markdown"}],"accordion":[{"title":"Background notes","markdown":"Notes markdown"}],"relatedData":[{"uri":"/economy/environmentalaccounts/datasets/ukenvironmentalaccountsenergybridging"}],"charts":[{"title":"Figure 1.1","filename":"38d8c337","uri":"/economy/environmentalaccounts/bulletins/ukenvironmentalaccounts/2015-07-09/38d8c337"}],"tables":[{"title":"Table 5.1","filename":"6f587872","uri":"/economy/environmentalaccounts/bulletins/ukenvironmentalaccounts/2015-07-09/6f587872"}],"images":[],"equations":[],"links":[],"type":"bulletin","uri":"/economy/environmentalaccounts/bulletins/ukenvironmentalaccounts/2015-07-09","description":{"title":"UK Environmental Accounts","summary":"Measures the contribution of the environment to the economy","keywords":["fuel, energy"],"metaDescription":"Measures the contribution of the environment.","nationalStatistic":true,"latestRelease":true,"contact":{"email":"environment.accounts@ons.gsi.gov.uk","name":"Someone","telephone":"+44 (0)1633 455680"},"releaseDate":"2015-07-08T23:00:00.000Z","nextRelease":"","edition":"2015","unit":"","preUnit":"","source":""}}`))
	case "/":
		w.Write([]byte(`{"intro":{"title":"Welcome to the Office for National Statistics","markdown":"Test markdown"},"featuredContent":[{"title":"Featured Content One","description":"Featured Content One Description","uri":"/one","image":"testImage"}],"aroundONS":[{"title":"Around ONS One","description":"Around ONS One Description","uri":"/one","image":"testImage"}],"serviceMessage":"","description":{"keywords":[ "keywordOne", "keywordTwo" ],"metaDescription":"","unit":"","preUnit":"","source":""}}`))
	case "notFound":
		w.WriteHeader(http.StatusNotFound)
	}

}

func parents(w http.ResponseWriter, req *http.Request) {
	log.Info(context.Background(), "+++++++client_test : parents...", log.Data{"uriString": ""})
	uri := req.URL.Query().Get("uri")

	switch uri {
	case "/employmentandlabourmarket/peopleinwork/workplacedisputesandworkingconditions/datasets/labourdisputesbysectorlabd02":
		w.Write([]byte(`[{"uri":"/","description":{"title":"Home"},"type":"home_page"},{"uri":"/employmentandlabourmarket","description":{"title":"Employment and labour market"},"type":"taxonomy_landing_page"},{"uri":"/employmentandlabourmarket/peopleinwork","description":{"title":"People in work"},"type":"taxonomy_landing_page"},{"uri":"/employmentandlabourmarket/peopleinwork/workplacedisputesandworkingconditions","description":{"title":"Workplace disputes and working conditions"},"type":"product_page"}]`))
	}
}

func filesize(w http.ResponseWriter, req *http.Request) {

	switch uri := req.URL.Query().Get("uri"); uri {
	case "filesize":
	case "12345/helloworld.txt":
	case "12345/test.txt":
	case "absoluteDownloadURI/test.txt":
	case "absoluteDownloadURI/helloworld.txt":
		break

	default:
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(errors.New("invalid path for get file size").Error()))
		return
	}

	zebedeeResponse := struct {
		FileSize int `json:"fileSize"`
	}{
		FileSize: 5242880,
	}

	b, err := json.Marshal(zebedeeResponse)
	if err != nil {
		w.Write([]byte(err.Error()))
		return
	}
	w.Write(b)
}

func TestUnitClient(t *testing.T) {
	portChan := make(chan int)
	go mockZebedeeServer(portChan)

	port := <-portChan
	cli := New(fmt.Sprintf("http://localhost:%d", port))

	ctx := context.Background()
	testCollectionID := "test-collection"

	Convey("test get()", t, func() {

		Convey("test get successfully returns response from zebedee with headers", func() {
			b, h, err := cli.get(ctx, testAccessToken, "/data?uri=foo")
			So(err, ShouldBeNil)
			So(len(h), ShouldEqual, 3)
			So(h.Get("Content-Length"), ShouldEqual, "2")
			So(h.Get("Content-Type"), ShouldEqual, "text/plain; charset=utf-8")
			So(h.Get("Date"), ShouldNotBeNil)
			So(string(b), ShouldEqual, `{}`)
		})

		Convey("test error returned if requesting invalid zebedee url", func() {
			b, h, err := cli.get(ctx, testAccessToken, "/invalid")
			So(err, ShouldNotBeNil)
			So(err, ShouldHaveSameTypeAs, ErrInvalidZebedeeResponse{})
			So(err.Error(), ShouldEqual, "invalid response from zebedee: 404, path: /invalid")
			So(b, ShouldBeNil)
			So(h, ShouldBeNil)
		})
	})

	Convey("test getLanding successfully returns a landing model", t, func() {
		m, err := cli.GetDatasetLandingPage(ctx, testAccessToken, "", "", "labour")
		So(err, ShouldBeNil)
		So(m, ShouldNotBeEmpty)
		So(m.Type, ShouldEqual, "dataset_landing_page")
	})

	Convey("GetHomepageContent() returns a homepage model", t, func() {
		m, err := cli.GetHomepageContent(ctx, testAccessToken, "", "", "/")
		So(err, ShouldBeNil)
		So(m, ShouldNotBeEmpty)
		So(m.Intro.Title, ShouldEqual, "Welcome to the Office for National Statistics")
		So(len(m.FeaturedContent), ShouldEqual, 1)
		So(m.FeaturedContent[0].Title, ShouldEqual, "Featured Content One")
		So(m.FeaturedContent[0].ImageID, ShouldEqual, "testImage")
		So(len(m.AroundONS), ShouldEqual, 1)
		So(m.AroundONS[0].Title, ShouldEqual, "Around ONS One")
		So(m.AroundONS[0].ImageID, ShouldEqual, "testImage")
		So(m.Description.Keywords[0], ShouldEqual, "keywordOne")
		So(m.ServiceMessage, ShouldEqual, "")
	})

	Convey("test get dataset details", t, func() {
		d, err := cli.GetDataset(ctx, testAccessToken, "", "", "12345")
		So(err, ShouldBeNil)
		So(d.URI, ShouldEqual, "www.google.com")
		So(d.SupplementaryFiles[0].Title, ShouldEqual, "helloworld")
	})

	Convey("test get dataset details with absolute url in download section", t, func() {
		d, err := cli.GetDataset(ctx, testAccessToken, "", "", "absoluteDownloadURI")
		So(err, ShouldBeNil)
		So(d.URI, ShouldEqual, "localhost")
		So(d.SupplementaryFiles[0].Title, ShouldEqual, "helloworld")
	})

	Convey("test getFileSize returns human readable filesize", t, func() {
		fs, err := cli.GetFileSize(ctx, testAccessToken, "", "", "filesize")
		So(err, ShouldBeNil)
		So(fs.Size, ShouldEqual, 5242880)
	})

	Convey("test getPageTitle returns a correctly formatted page title", t, func() {
		t, err := cli.GetPageTitle(ctx, testAccessToken, "", "", "pageTitle")
		So(err, ShouldBeNil)
		So(t.Title, ShouldEqual, "baby-names")
		So(t.Edition, ShouldEqual, "2017")
	})

	Convey("test createRequestURL", t, func() {
		Convey("test collection ID is added to URL when collection ID is passed", func() {
			url := cli.createRequestURL(ctx, testCollectionID, "", "/data", "uri=/test/path/123")
			So(url, ShouldEqual, "/data/test-collection?uri=%2Ftest%2Fpath%2F123")
		})
		Convey("test collection ID is not added to URL when empty collection ID is passed", func() {
			url := cli.createRequestURL(ctx, "", "", "/data", "uri=/test/path/123")
			So(url, ShouldEqual, "/data?uri=%2Ftest%2Fpath%2F123")
		})
		Convey("test lang query parameter is added to URL when locale code is passed", func() {
			url := cli.createRequestURL(ctx, "", "cy", "/data", "uri=/test/path/123")
			So(url, ShouldEqual, "/data?uri=%2Ftest%2Fpath%2F123&lang=cy")
		})
		Convey("test collection ID and lang query parameter are added to URL when collection ID and locale code are present", func() {
			url := cli.createRequestURL(ctx, testCollectionID, "cy", "/data", "uri=/test/path/123")
			So(url, ShouldEqual, "/data/test-collection?uri=%2Ftest%2Fpath%2F123&lang=cy")
		})
	})

	Convey("test GetBulletin", t, func() {
		Convey("returns a bulletin", func() {
			b, err := cli.GetBulletin(ctx, testAccessToken, "", "bulletin")
			So(err, ShouldBeNil)
			So(b, ShouldNotBeEmpty)
			So(b.Type, ShouldEqual, "bulletin")
			So(b.URI, ShouldEqual, "/economy/environmentalaccounts/bulletins/ukenvironmentalaccounts/2015-07-09")
			So(b.RelatedBulletins, ShouldNotBeEmpty)
			So(len(b.RelatedBulletins), ShouldEqual, 1)
			So(b.RelatedBulletins[0].URI, ShouldEqual, "/economy/environmentalaccounts/datasets/ukenvironmentalaccountsenergybridging")
			So(b.Sections, ShouldNotBeEmpty)
			So(len(b.Sections), ShouldEqual, 2)
			So(b.Sections[0].Title, ShouldEqual, "Main points")
			So(b.Sections[0].Markdown, ShouldEqual, "Main points markdown")
			So(b.Sections[1].Title, ShouldEqual, "Overview")
			So(b.Sections[1].Markdown, ShouldEqual, "Overview markdown")
			So(b.Accordion, ShouldNotBeEmpty)
			So(len(b.Accordion), ShouldEqual, 1)
			So(b.Accordion[0].Title, ShouldEqual, "Background notes")
			So(b.Accordion[0].Markdown, ShouldEqual, "Notes markdown")
			So(b.RelatedData, ShouldNotBeEmpty)
			So(len(b.RelatedData), ShouldEqual, 1)
			So(b.RelatedData[0].URI, ShouldEqual, "/economy/environmentalaccounts/datasets/ukenvironmentalaccountsenergybridging")
			So(b.Charts, ShouldNotBeEmpty)
			So(len(b.Charts), ShouldEqual, 1)
			So(b.Charts[0].Title, ShouldEqual, "Figure 1.1")
			So(b.Charts[0].Filename, ShouldEqual, "38d8c337")
			So(b.Charts[0].URI, ShouldEqual, "/economy/environmentalaccounts/bulletins/ukenvironmentalaccounts/2015-07-09/38d8c337")
			So(b.Tables, ShouldNotBeEmpty)
			So(len(b.Tables), ShouldEqual, 1)
			So(b.Tables[0].Title, ShouldEqual, "Table 5.1")
			So(b.Tables[0].Filename, ShouldEqual, "6f587872")
			So(b.Tables[0].URI, ShouldEqual, "/economy/environmentalaccounts/bulletins/ukenvironmentalaccounts/2015-07-09/6f587872")
			So(b.Description, ShouldNotBeEmpty)
			So(b.Description.Title, ShouldEqual, "UK Environmental Accounts")
			So(b.Description.Summary, ShouldEqual, "Measures the contribution of the environment to the economy")
			So(b.Description.Title, ShouldEqual, "UK Environmental Accounts")
			So(b.Description.MetaDescription, ShouldEqual, "Measures the contribution of the environment.")
			So(b.Description.NationalStatistic, ShouldBeTrue)
			So(b.Description.Edition, ShouldEqual, "2015")
			So(b.Description.ReleaseDate, ShouldEqual, "2015-07-08T23:00:00.000Z")
			So(b.Description.Contact, ShouldNotBeEmpty)
			So(b.Description.Contact.Email, ShouldEqual, "environment.accounts@ons.gsi.gov.uk")
			So(b.Description.Contact.Name, ShouldEqual, "Someone")
			So(b.Description.Contact.Telephone, ShouldEqual, "+44 (0)1633 455680")
		})

		Convey("returns an error if uri not found", func() {
			b, err := cli.GetBulletin(ctx, testAccessToken, "", "notFound")
			So(err, ShouldNotBeNil)
			So(b, ShouldResemble, Bulletin{})
		})
	})
}

func TestClient_HealthChecker(t *testing.T) {
	ctx := context.Background()
	timePriorHealthCheck := time.Now()
	path := "/health"

	Convey("given clienter.Do returns an error", t, func() {
		clientError := errors.New("disciples of the watch obey")
		httpClient := newMockHTTPClient(&http.Response{}, clientError)
		zebedeeClient := newZebedeeClient(httpClient)
		check := initialState

		Convey("when zebedeeClient.Checker is called", func() {
			err := zebedeeClient.Checker(ctx, &check)
			So(err, ShouldBeNil)

			Convey("then the expected check is returned", func() {
				So(check.Name(), ShouldEqual, service)
				So(check.Status(), ShouldEqual, healthcheck.StatusCritical)
				So(check.StatusCode(), ShouldEqual, 0)
				So(check.Message(), ShouldEqual, clientError.Error())
				So(*check.LastChecked(), ShouldHappenAfter, timePriorHealthCheck)
				So(check.LastSuccess(), ShouldBeNil)
				So(*check.LastFailure(), ShouldHappenAfter, timePriorHealthCheck)
			})

			Convey("and client.Do should be called once with the expected parameters", func() {
				doCalls := httpClient.DoCalls()
				So(doCalls, ShouldHaveLength, 1)
				So(doCalls[0].Req.URL.Path, ShouldEqual, path)
			})
		})
	})

	Convey("given a 500 response", t, func() {
		httpClient := newMockHTTPClient(&http.Response{StatusCode: http.StatusInternalServerError}, nil)
		zebedeeClient := newZebedeeClient(httpClient)
		check := initialState

		Convey("when zebedeeClient.Checker is called", func() {
			err := zebedeeClient.Checker(ctx, &check)
			So(err, ShouldBeNil)

			Convey("then the expected check is returned", func() {
				So(check.Name(), ShouldEqual, service)
				So(check.Status(), ShouldEqual, healthcheck.StatusCritical)
				So(check.StatusCode(), ShouldEqual, 500)
				So(check.Message(), ShouldEqual, service+health.StatusMessage[healthcheck.StatusCritical])
				So(*check.LastChecked(), ShouldHappenAfter, timePriorHealthCheck)
				So(check.LastSuccess(), ShouldBeNil)
				So(*check.LastFailure(), ShouldHappenAfter, timePriorHealthCheck)
			})

			Convey("and client.Do should be called once with the expected parameters", func() {
				doCalls := httpClient.DoCalls()
				So(doCalls, ShouldHaveLength, 1)
				So(doCalls[0].Req.URL.Path, ShouldEqual, path)
			})
		})
	})

	Convey("given a 404 response", t, func() {
		httpClient := newMockHTTPClient(&http.Response{StatusCode: http.StatusNotFound}, nil)
		zebedeeClient := newZebedeeClient(httpClient)
		check := initialState

		Convey("when zebedeeClient.Checker is called", func() {
			err := zebedeeClient.Checker(ctx, &check)
			So(err, ShouldBeNil)

			Convey("then the expected check is returned", func() {
				So(check.Name(), ShouldEqual, service)
				So(check.Status(), ShouldEqual, healthcheck.StatusCritical)
				So(check.StatusCode(), ShouldEqual, 404)
				So(check.Message(), ShouldEqual, service+health.StatusMessage[healthcheck.StatusCritical])
				So(*check.LastChecked(), ShouldHappenAfter, timePriorHealthCheck)
				So(check.LastSuccess(), ShouldBeNil)
				So(*check.LastFailure(), ShouldHappenAfter, timePriorHealthCheck)
			})

			Convey("and client.Do should be called once with the expected parameters", func() {
				doCalls := httpClient.DoCalls()
				So(doCalls, ShouldHaveLength, 2)
				So(doCalls[0].Req.URL.Path, ShouldEqual, path)
				So(doCalls[1].Req.URL.Path, ShouldEqual, "/healthcheck")
			})
		})
	})

	Convey("given a 429 response", t, func() {
		httpClient := newMockHTTPClient(&http.Response{StatusCode: http.StatusTooManyRequests}, nil)
		httpClient.SetPathsWithNoRetries([]string{path, "/healthcheck"})

		zebedeeClient := newZebedeeClient(httpClient)
		check := initialState

		Convey("when zebedeeClient.Checker is called", func() {
			err := zebedeeClient.Checker(ctx, &check)
			So(err, ShouldBeNil)

			Convey("then the expected check is returned", func() {
				So(check.Name(), ShouldEqual, service)
				So(check.Status(), ShouldEqual, healthcheck.StatusWarning)
				So(check.StatusCode(), ShouldEqual, 429)
				So(check.Message(), ShouldEqual, service+health.StatusMessage[healthcheck.StatusWarning])
				So(*check.LastChecked(), ShouldHappenAfter, timePriorHealthCheck)
				So(check.LastSuccess(), ShouldBeNil)
				So(*check.LastFailure(), ShouldHappenAfter, timePriorHealthCheck)
			})

			Convey("and client.Do should be called once with the expected parameters", func() {
				doCalls := httpClient.DoCalls()
				So(doCalls, ShouldHaveLength, 1)
				So(doCalls[0].Req.URL.Path, ShouldEqual, path)
			})
		})
	})

	Convey("given a 200 response", t, func() {
		httpClient := newMockHTTPClient(&http.Response{StatusCode: http.StatusOK}, nil)
		zebedeeClient := newZebedeeClient(httpClient)
		check := initialState

		Convey("when zebedeeClient.Checker is called", func() {
			err := zebedeeClient.Checker(ctx, &check)
			So(err, ShouldBeNil)

			Convey("then the expected check is returned", func() {
				So(check.Name(), ShouldEqual, service)
				So(check.Status(), ShouldEqual, healthcheck.StatusOK)
				So(check.StatusCode(), ShouldEqual, 200)
				So(check.Message(), ShouldEqual, service+health.StatusMessage[healthcheck.StatusOK])
				So(*check.LastChecked(), ShouldHappenAfter, timePriorHealthCheck)
				So(*check.LastSuccess(), ShouldHappenAfter, timePriorHealthCheck)
				So(check.LastFailure(), ShouldBeNil)
			})

			Convey("and client.Do should be called once with the expected parameters", func() {
				doCalls := httpClient.DoCalls()
				So(doCalls, ShouldHaveLength, 1)
				So(doCalls[0].Req.URL.Path, ShouldEqual, path)
			})
		})
	})
}

func TestClient_PublishedDataEndpoint(t *testing.T) {

	ctx := context.Background()
	path := "/publisheddata"
	testURIString := "/employmentandlabourmarket/peopleinwork/workplacedisputesandworkingconditions/datasets/labourdisputesbysectorlabd02"
	documentContent := []byte("{byte slice returned}")
	body := httpmocks.NewReadCloserMock(documentContent, nil)

	Convey("given a 200 response", t, func() {
		response := httpmocks.NewResponseMock(body, http.StatusOK)
		httpClient := newMockHTTPClient(response, nil)
		zebedeeClient := newZebedeeClient(httpClient)

		Convey("when zebedeeClient.GetPublised is called", func() {
			testContent, err := zebedeeClient.GetPublishedData(ctx, testURIString)

			Convey("then the expected content is returned", func() {
				So(err, ShouldBeNil)
				So(testContent, ShouldNotBeNil)
				So(testContent, ShouldResemble, documentContent)
			})

			Convey("and client.Do should be called once with the expected parameters", func() {
				doCalls := httpClient.DoCalls()
				So(doCalls, ShouldHaveLength, 1)
				So(doCalls[0].Req.URL.Path, ShouldEqual, path)
				p := doCalls[0].Req.FormValue("uri")
				So(p, ShouldEqual, testURIString)
			})
		})
	})
	Convey("given a 500 response", t, func() {
		response := httpmocks.NewResponseMock(body, http.StatusInternalServerError)
		httpClient := newMockHTTPClient(response, nil)
		zebedeeClient := newZebedeeClient(httpClient)

		Convey("when zebedeeClient.GetPublishedData is called", func() {
			testContent, err := zebedeeClient.GetPublishedData(ctx, testURIString)

			Convey("then the expected error is returned", func() {
				So(err, ShouldNotBeNil)
				So(testContent, ShouldBeNil)
				So(err, ShouldHaveSameTypeAs, ErrInvalidZebedeeResponse{})
				So(err.Error(), ShouldEqual, "invalid response from zebedee: 500, path: /publisheddata")

			})

			Convey("and client.Do should be called once with the expected parameters", func() {
				doCalls := httpClient.DoCalls()
				So(doCalls, ShouldHaveLength, 1)
				So(doCalls[0].Req.URL.Path, ShouldEqual, path)
			})
		})
	})
	Convey("given a 404 response", t, func() {
		response := httpmocks.NewResponseMock(body, http.StatusNotFound)
		httpClient := newMockHTTPClient(response, nil)
		zebedeeClient := newZebedeeClient(httpClient)

		Convey("when zebedeeClient.GetPublishedData is called", func() {
			testContent, err := zebedeeClient.GetPublishedData(ctx, testURIString)

			Convey("then the expected error is returned", func() {
				So(err, ShouldNotBeNil)
				So(testContent, ShouldBeNil)
				So(err, ShouldHaveSameTypeAs, ErrInvalidZebedeeResponse{})
				So(err.Error(), ShouldEqual, "invalid response from zebedee: 404, path: /publisheddata")

			})

			Convey("and client.Do should be called once with the expected parameters", func() {
				doCalls := httpClient.DoCalls()
				So(doCalls, ShouldHaveLength, 1)
				So(doCalls[0].Req.URL.Path, ShouldEqual, path)
			})
		})
	})
}
