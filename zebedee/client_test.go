package zebedee

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/ONSdigital/dp-api-clients-go/v2/health"
	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	"github.com/ONSdigital/dp-mocking/httpmocks"
	dphttp "github.com/ONSdigital/dp-net/v2/http"
	dprequest "github.com/ONSdigital/dp-net/v2/request"
	"github.com/ONSdigital/log.go/v2/log"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	. "github.com/smartystreets/goconvey/convey"
)

const testHost = "http://localhost:8080"

var (
	testAccessToken        = "test-access-token"
	testCollectionID       = "test-collection"
	testFileSize           = 5242880
	testFileSizeCollection = 3313490
	testLang               = "en"
	initialState           = health.CreateCheckState(service)
)

func mockZebedeeServer(port chan int) {
	r := mux.NewRouter()

	r.Path("/data").HandlerFunc(contentData)
	r.Path("/data/{collectionID}").HandlerFunc(contentDataCollection)
	r.Path("/filesize").HandlerFunc(filesize)
	r.Path("/filesize/{collectionID}").HandlerFunc(filesizeCollection)

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

	checkAccessToken(w, req)
	checkLanguage(w, req)

	switch uri {
	case "foo":
		w.Write([]byte(`{}`))
	case "labour":
		w.Write([]byte(`{"downloads":[{"title":"Latest","file":"/employmentandlabourmarket/peopleinwork/workplacedisputesandworkingconditions/datasets/labourdisputesbysectorlabd02/labd02jul2015_tcm77-408195.xls"}],"section":{"markdown":""},"relatedDatasets":[{"uri":"/employmentandlabourmarket/peopleinwork/workplacedisputesandworkingconditions/datasets/labourdisputeslabd01"},{"uri":"/employmentandlabourmarket/peopleinwork/workplacedisputesandworkingconditions/datasets/stoppagesofworklabd03"}],"relatedDocuments":[{"uri":"/employmentandlabourmarket/peopleinwork/employmentandemployeetypes/bulletins/uklabourmarket/2015-07-15"}],"relatedMethodology":[],"type":"dataset_landing_page","uri":"/employmentandlabourmarket/peopleinwork/workplacedisputesandworkingconditions/datasets/labourdisputesbysectorlabd02","description":{"title":"Labour disputes by sector: LABD02","summary":"Labour disputes by sector.","keywords":["strike"],"metaDescription":"Labour disputes by sector.","nationalStatistic":true,"contact":{"email":"richard.clegg@ons.gsi.gov.uk\n","name":"Richard Clegg\n","telephone":"+44 (0)1633 455400 \n"},"releaseDate":"2015-07-14T23:00:00.000Z","nextRelease":"12 August 2015","datasetId":"","unit":"","preUnit":"","source":""}}`))
	case "dataset":
		w.Write([]byte(`{"type":"dataset","uri":"www.google.com","downloads":[{"file":"test.txt"}],"supplementaryFiles":[{"title":"helloworld","file":"helloworld.txt"}],"versions":[{"uri":"www.google.com"}]}`))
	case "pageTitle1":
		w.Write([]byte(`{"title":"baby-names","edition":"2017","uri":"path/to/baby-names/2017"}`))
	case "pageTitle2":
		w.Write([]byte(`{"title":"page-title","edition":"2021","uri":"path/to/page-title/2021"}`))
	case "pageDescription1":
		w.Write([]byte(`{"uri":"path/to/page-description","description":{"title":"Page title", "summary":"This is the page summary","keywords":["Economy","Retail"],"metaDescription":"meta","nationalStatistic":true,"latestRelease":true,"contact":{"email": "contact@ons.gov.uk","name":"Contact","telephone":"+44 (0) 1633 456900"},"releaseDate":"2015-09-14T23:00:00.000Z","nextRelease":"13 October 2015","edition":"August 2015"}}`))
	case "pageDescription2":
		w.Write([]byte(`{"uri":"page-description-2","description":{"title":"UK Environmental Accounts", "summary":"Measuring the contribution of the environment to the economy","keywords":["emissions","climate"],"metaDescription":"meta2","nationalStatistic":true,"latestRelease":true,"contact":{"email": "contact@ons.gov.uk","name":"Contact","telephone":"+44 (0) 1633 456900"},"releaseDate":"2021-06-02T23:00:00.000Z","nextRelease":"June 2022","edition":"2021"}}`))
	case "pageDescription3":
		w.Write([]byte(`{"uri":"page-description-3","description":{"title":"Another page title", "summary":"","_abstract":"Page description is mapped from _abstract"}}`))
	case "bulletin-latest-release":
		w.Write([]byte(`{"relatedBulletins":[{"uri":"pageTitle1"}],"sections":[{"title":"Main points","markdown":"Main points markdown"},{"title":"Overview","markdown":"Overview markdown"}],"accordion":[{"title":"Background notes","markdown":"Notes markdown"}],"relatedData":[{"uri":"/economy/environmentalaccounts/datasets/ukenvironmentalaccountsenergybridging"}],"charts":[{"title":"Figure 1.1","filename":"38d8c337","uri":"/economy/environmentalaccounts/bulletins/ukenvironmentalaccounts/2015-07-09/38d8c337"}],"tables":[{"title":"Table 5.1","filename":"6f587872","uri":"/economy/environmentalaccounts/bulletins/ukenvironmentalaccounts/2015-07-09/6f587872"}],"images":[],"equations":[],"links":[{"uri":"pageTitle1"}, {"uri":"pageTitle2"}],"alerts":[{"date":"2021-09-30T07:10:46.230Z","markdown":"alert"}],"versions":[{"uri":"v1","updateDate":"2021-10-19T10:43:34.507Z","correctionNotice":"Notice"}],"type":"bulletin","uri":"/bulletin/2015-07-09","description":{"title":"UK Environmental Accounts","summary":"Measures the contribution of the environment to the economy","keywords":["fuel, energy"],"metaDescription":"Measures the contribution of the environment.","nationalStatistic":true,"latestRelease":true,"contact":{"email":"environment.accounts@ons.gsi.gov.uk","name":"Someone","telephone":"+44 (0)1633 455680"},"releaseDate":"2015-07-08T23:00:00.000Z","nextRelease":"","edition":"2015","unit":"","preUnit":"","source":""}}`))
	case "bulletin-not-latest-release":
		w.Write([]byte(`{"relatedBulletins":[{"uri":"pageTitle1"}],"sections":[{"title":"Main points","markdown":"Main points markdown"},{"title":"Overview","markdown":"Overview markdown"}],"accordion":[{"title":"Background notes","markdown":"Notes markdown"}],"relatedData":[{"uri":"/economy/environmentalaccounts/datasets/ukenvironmentalaccountsenergybridging"}],"charts":[{"title":"Figure 1.1","filename":"38d8c337","uri":"/economy/environmentalaccounts/bulletins/ukenvironmentalaccounts/2015-07-09/38d8c337"}],"tables":[{"title":"Table 5.1","filename":"6f587872","uri":"/economy/environmentalaccounts/bulletins/ukenvironmentalaccounts/2015-07-09/6f587872"}],"images":[],"equations":[],"links":[{"uri":"pageTitle1"}, {"uri":"pageTitle2"}],"alerts":[{"date":"2021-09-30T07:10:46.230Z","markdown":"alert"}],"versions":[{"uri":"v1","updateDate":"2021-10-19T10:43:34.507Z","correctionNotice":"Notice"}],"type":"bulletin","uri":"/bulletin/2015-07-09","description":{"title":"UK Environmental Accounts","summary":"Measures the contribution of the environment to the economy","keywords":["fuel, energy"],"metaDescription":"Measures the contribution of the environment.","nationalStatistic":true,"latestRelease":false,"contact":{"email":"environment.accounts@ons.gsi.gov.uk","name":"Someone","telephone":"+44 (0)1633 455680"},"releaseDate":"2015-07-08T23:00:00.000Z","nextRelease":"","edition":"2015","unit":"","preUnit":"","source":""}}`))
	case "/bulletin/latest":
		w.Write([]byte(`{"title":"latest release","edition":"2021","uri":"/bulletin/collection/2021"}`))
	case "/release":
		w.Write([]byte(`{"markdown":["markdown"],"relatedDocuments":[{"uri":"pageDescription2"}],"relatedDatasets":[{"uri":"pageDescription1"}],"relatedAPIDatasets":[{"uri":"cantabularDataset","title":"Title for cantabularDataset"},{"uri":"cmdDataset","title":"Title for cmdDataset"}],"relatedMethodology":[{"uri":"pageDescription1"}],"relatedMethodologyArticle":[{"uri":"pageDescription2"}],"links":[{"uri":"pageDescription1"}, {"uri":"pageDescription2"}, {"uri":"externalLinkURI","title":"This is a link to an external site"}],"dateChanges":[{"previousDate":"2021-08-15T11:12:05.592Z","changeNotice":"change notice"}],"uri":"/releases/indexofproductionukdecember2021timeseries","description":{"finalised":true,"title":"Index of Production","summary":"Movements in the volume of production for the UK production industries","nationalStatistic":true,"contact":{"email":"indexofproduction@ons.gov.uk","name":"Contact name","telephone":"+44 1633 456980"},"releaseDate":"2022-02-11T07:00:00.000Z","nextRelease":"11 March 2022","cancelled":true,"cancellationNotice":["notice"],"finalised":true,"published":true,"provisionalDate":"Dec 22"}}`))
	case "/":
		w.Write([]byte(`{"intro":{"title":"Welcome to the Office for National Statistics","markdown":"Test markdown"},"featuredContent":[{"title":"Featured Content One","description":"Featured Content One Description","uri":"/one","image":"testImage"}],"aroundONS":[{"title":"Around ONS One","description":"Around ONS One Description","uri":"/one","image":"testImage"}],"serviceMessage":"","emergencyBanner":{"type":"notable_death","title":"Emergency banner title","description":"Emergency banner description","uri":"www.google.com","linkText":"More info"},"description":{"keywords":[ "keywordOne", "keywordTwo" ],"metaDescription":"","unit":"","preUnit":"","source":""}}`))
	case "notFound":
		w.WriteHeader(http.StatusNotFound)
	}

}

func contentDataCollection(w http.ResponseWriter, req *http.Request) {
	uri := req.URL.Query().Get("uri")

	checkAccessToken(w, req)
	checkLanguage(w, req)
	checkCollection(w, req)

	switch uri {
	case "labour":
		w.Write([]byte(`{"downloads":[{"title":"Latest","file":"/employmentandlabourmarket/peopleinwork/workplacedisputesandworkingconditions/datasets/labourdisputesbysectorlabd02/labd02jul2015_tcm77-408195.xls"}],"section":{"markdown":""},"relatedDatasets":[{"uri":"/employmentandlabourmarket/peopleinwork/workplacedisputesandworkingconditions/datasets/labourdisputeslabd01"},{"uri":"/employmentandlabourmarket/peopleinwork/workplacedisputesandworkingconditions/datasets/stoppagesofworklabd03"}],"relatedDocuments":[{"uri":"/employmentandlabourmarket/peopleinwork/employmentandemployeetypes/bulletins/uklabourmarket/2015-07-15"}],"relatedMethodology":[],"type":"dataset_landing_page","uri":"/employmentandlabourmarket/peopleinwork/workplacedisputesandworkingconditions/datasets/labourdisputesbysectorlabd02","description":{"title":"Labour disputes by sector: LABD02 - publishing","summary":"Labour disputes by sector.","keywords":["strike"],"metaDescription":"Labour disputes by sector.","nationalStatistic":true,"contact":{"email":"richard.clegg@ons.gsi.gov.uk\n","name":"Richard Clegg\n","telephone":"+44 (0)1633 455400 \n"},"releaseDate":"2015-07-14T23:00:00.000Z","nextRelease":"12 August 2015","datasetId":"","unit":"","preUnit":"","source":""}}`))
	case "dataset":
		w.Write([]byte(`{"type":"dataset","uri":"www.google.com","downloads":[{"file":"testCollection.txt"}],"supplementaryFiles":[{"title":"helloworld","file":"helloworld.txt"}],"versions":[{"uri":"www.google.com"}]}`))
	case "pageTitle1":
		w.Write([]byte(`{"title":"baby-names","edition":"collection","uri":"path/to/baby-names/collection"}`))
	case "pageTitle2":
		w.Write([]byte(`{"title":"page-title","edition":"c2021","uri":"path/to/page-title/2021"}`))
	case "pageDescription1":
		w.Write([]byte(`{"uri":"path/to/page-description/collection","description":{"title":"Page title", "summary":"This is the page summary","keywords":["Economy","Retail"],"metaDescription":"meta","nationalStatistic":true,"latestRelease":true,"contact":{"email": "contact@ons.gov.uk","name":"Contact","telephone":"+44 (0) 1633 456900"},"releaseDate":"2015-09-14T23:00:00.000Z","nextRelease":"13 October 2015","edition":"collection"}}`))
	case "pageDescription2":
		w.Write([]byte(`{"uri":"collection/page-description-2","description":{"title":"Collection UK Environmental Accounts", "summary":"Measuring the contribution of the environment to the economy","keywords":["emissions","climate"],"metaDescription":"meta2","nationalStatistic":true,"latestRelease":true,"contact":{"email": "contact@ons.gov.uk","name":"Contact","telephone":"+44 (0) 1633 456900"},"releaseDate":"2021-06-02T23:00:00.000Z","nextRelease":"June 2022","edition":"2021c"}}`))
	case "pageDescription3":
		w.Write([]byte(`{"uri":"collection/page-description-3","description":{"title":"Another page title", "summary":"", "_abstract": "Summary is from the _abstract field"}}`))
	case "bulletin-latest-release":
		w.Write([]byte(`{"relatedBulletins":[{"uri":"pageTitle1"}],"sections":[{"title":"Main points","markdown":"Main points markdown"},{"title":"Overview","markdown":"Overview markdown"}],"accordion":[{"title":"Background notes","markdown":"Notes markdown"}],"relatedData":[{"uri":"/economy/environmentalaccounts/datasets/ukenvironmentalaccountsenergybridging"}],"charts":[{"title":"Figure 1.1","filename":"38d8c337","uri":"/economy/environmentalaccounts/bulletins/ukenvironmentalaccounts/2015-07-09/38d8c337"}],"tables":[{"title":"Table 5.1","filename":"6f587872","uri":"/economy/environmentalaccounts/bulletins/ukenvironmentalaccounts/2015-07-09/6f587872"}],"images":[],"equations":[],"links":[{"uri":"pageTitle1"}, {"uri":"pageTitle2"}],"alerts":[{"date":"2021-09-30T07:10:46.230Z","markdown":"alert"}],"versions":[{"uri":"v1","updateDate":"2021-10-19T10:43:34.507Z","correctionNotice":"Notice"}],"type":"bulletin","uri":"/bulletin/2015-07-09","description":{"title":"UK Environmental Accounts with collection","summary":"Measures the contribution of the environment to the economy","keywords":["fuel, energy"],"metaDescription":"Measures the contribution of the environment.","nationalStatistic":true,"latestRelease":true,"contact":{"email":"environment.accounts@ons.gsi.gov.uk","name":"Someone","telephone":"+44 (0)1633 455680"},"releaseDate":"2015-07-08T23:00:00.000Z","nextRelease":"","edition":"2015","unit":"","preUnit":"","source":""}}`))
	case "bulletin-not-latest-release":
		w.Write([]byte(`{"relatedBulletins":[{"uri":"pageTitle1"}],"sections":[{"title":"Main points","markdown":"Main points markdown"},{"title":"Overview","markdown":"Overview markdown"}],"accordion":[{"title":"Background notes","markdown":"Notes markdown"}],"relatedData":[{"uri":"/economy/environmentalaccounts/datasets/ukenvironmentalaccountsenergybridging"}],"charts":[{"title":"Figure 1.1","filename":"38d8c337","uri":"/economy/environmentalaccounts/bulletins/ukenvironmentalaccounts/2015-07-09/38d8c337"}],"tables":[{"title":"Table 5.1","filename":"6f587872","uri":"/economy/environmentalaccounts/bulletins/ukenvironmentalaccounts/2015-07-09/6f587872"}],"images":[],"equations":[],"links":[{"uri":"pageTitle1"}, {"uri":"pageTitle2"}],"alerts":[{"date":"2021-09-30T07:10:46.230Z","markdown":"alert"}],"versions":[{"uri":"v1","updateDate":"2021-10-19T10:43:34.507Z","correctionNotice":"Notice"}],"type":"bulletin","uri":"/bulletin/2015-07-09","description":{"title":"UK Environmental Accounts with collection","summary":"Measures the contribution of the environment to the economy","keywords":["fuel, energy"],"metaDescription":"Measures the contribution of the environment.","nationalStatistic":true,"latestRelease":false,"contact":{"email":"environment.accounts@ons.gsi.gov.uk","name":"Someone","telephone":"+44 (0)1633 455680"},"releaseDate":"2015-07-08T23:00:00.000Z","nextRelease":"","edition":"2015","unit":"","preUnit":"","source":""}}`))
	case "/bulletin/latest":
		w.Write([]byte(`{"title":"latest release","edition":"2021","uri":"/bulletin/2021"}`))
	case "/release":
		w.Write([]byte(`{"markdown":["collection markdown"],"relatedDocuments":[{"uri":"pageDescription2"},{"uri":"pageDescription3"}],"relatedDatasets":[{"uri":"pageDescription1"}],"relatedMethodology":[{"uri":"pageDescription1"}],"relatedMethodologyArticle":[{"uri":"pageDescription2"}],"links":[{"uri":"pageDescription1"}, {"uri":"pageDescription2"}, {"uri":"externalLinkURI","title":"This is a link to an external site"}],"dateChanges":[{"previousDate":"2021-08-15T11:12:05.592Z","changeNotice":"change notice"}],"uri":"/releases/collection","description":{"finalised":true,"title":"Index of Production","summary":"Movements in the volume of production for the UK production industries","nationalStatistic":true,"contact":{"email":"indexofproduction@ons.gov.uk","name":"Contact name","telephone":"+44 1633 456980"},"releaseDate":"2022-02-11T07:00:00.000Z","nextRelease":"11 March 2022","cancelled":true,"cancellationNotice":["notice"],"finalised":true,"published":true,"provisionalDate":"Dec 22"}}`))
	case "/":
		w.Write([]byte(`{"intro":{"title":"Welcome to Publishing","markdown":"Test markdown"},"featuredContent":[{"title":"Featured Content One","description":"Featured Content One Description","uri":"/one","image":"testImage"}],"aroundONS":[{"title":"Around ONS One","description":"Around ONS One Description","uri":"/one","image":"testImage"}],"serviceMessage":"","emergencyBanner":{"type":"notable_death","title":"Emergency banner title","description":"Emergency banner description","uri":"www.google.com","linkText":"More info"},"description":{"keywords":[ "keywordOne", "keywordTwo" ],"metaDescription":"","unit":"","preUnit":"","source":""}}`))
	}
}

func checkAccessToken(w http.ResponseWriter, req *http.Request) {
	serviceAuthToken := req.Header.Get(dprequest.FlorenceHeaderKey)
	if serviceAuthToken != testAccessToken {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("401 - No access token header set!"))
	}
}

func checkCollection(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	collectionID := vars["collectionID"]
	if collectionID != testCollectionID {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("400 - Wrong collection id set!"))
	}
}

func checkLanguage(w http.ResponseWriter, req *http.Request) {
	lang := req.URL.Query().Get("lang")
	if lang != testLang {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("400 - Wrong language!"))
	}
}

func filesize(w http.ResponseWriter, req *http.Request) {
	writeFilesizeResponse(w, testFileSize)
}

func filesizeCollection(w http.ResponseWriter, req *http.Request) {
	checkCollection(w, req)
	writeFilesizeResponse(w, testFileSizeCollection)
}

func writeFilesizeResponse(w http.ResponseWriter, filesize int) {
	zebedeeResponse := struct {
		FileSize int `json:"fileSize"`
	}{
		FileSize: filesize,
	}

	b, err := json.Marshal(zebedeeResponse)
	if err != nil {
		w.Write([]byte(err.Error()))
		return
	}
	w.Write(b)
}

func TestUnitClient(t *testing.T) {
	t.Parallel()
	portChan := make(chan int)
	go mockZebedeeServer(portChan)

	port := <-portChan
	cli := New(fmt.Sprintf("http://localhost:%d", port))

	ctx := context.Background()

	Convey("test get()", t, func() {

		Convey("test get successfully returns response from zebedee with headers", func() {
			b, h, err := cli.get(ctx, testAccessToken, "/data?uri=foo&lang="+testLang)
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
		m, err := cli.GetDatasetLandingPage(ctx, testAccessToken, "", testLang, "labour")
		So(err, ShouldBeNil)
		So(m, ShouldNotBeEmpty)
		So(m.Type, ShouldEqual, "dataset_landing_page")
		So(m.Description.Title, ShouldEqual, "Labour disputes by sector: LABD02")
	})

	Convey("test getLanding successfully returns a landing model when using a collection", t, func() {
		m, err := cli.GetDatasetLandingPage(ctx, testAccessToken, testCollectionID, testLang, "labour")
		So(err, ShouldBeNil)
		So(m, ShouldNotBeEmpty)
		So(m.Type, ShouldEqual, "dataset_landing_page")
		So(m.Description.Title, ShouldEqual, "Labour disputes by sector: LABD02 - publishing")
	})

	Convey("GetHomepageContent() returns a homepage model", t, func() {
		m, err := cli.GetHomepageContent(ctx, testAccessToken, "", testLang, "/")
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
		So(m.EmergencyBanner.Title, ShouldEqual, "Emergency banner title")
		So(m.EmergencyBanner.Type, ShouldEqual, "notable_death")
		So(m.EmergencyBanner.Description, ShouldEqual, "Emergency banner description")
		So(m.EmergencyBanner.URI, ShouldEqual, "www.google.com")
		So(m.EmergencyBanner.LinkText, ShouldEqual, "More info")
	})

	Convey("GetHomepageContent() returns a homepage model when using a collection", t, func() {
		m, err := cli.GetHomepageContent(ctx, testAccessToken, testCollectionID, testLang, "/")
		So(err, ShouldBeNil)
		So(m, ShouldNotBeEmpty)
		So(m.Intro.Title, ShouldEqual, "Welcome to Publishing")
		So(len(m.FeaturedContent), ShouldEqual, 1)
		So(m.FeaturedContent[0].Title, ShouldEqual, "Featured Content One")
		So(m.FeaturedContent[0].ImageID, ShouldEqual, "testImage")
		So(len(m.AroundONS), ShouldEqual, 1)
		So(m.AroundONS[0].Title, ShouldEqual, "Around ONS One")
		So(m.AroundONS[0].ImageID, ShouldEqual, "testImage")
		So(m.Description.Keywords[0], ShouldEqual, "keywordOne")
		So(m.ServiceMessage, ShouldEqual, "")
		So(m.EmergencyBanner.Title, ShouldEqual, "Emergency banner title")
		So(m.EmergencyBanner.Type, ShouldEqual, "notable_death")
		So(m.EmergencyBanner.Description, ShouldEqual, "Emergency banner description")
		So(m.EmergencyBanner.URI, ShouldEqual, "www.google.com")
		So(m.EmergencyBanner.LinkText, ShouldEqual, "More info")
	})

	Convey("test get dataset details", t, func() {
		d, err := cli.GetDataset(ctx, testAccessToken, "", testLang, "dataset")
		So(err, ShouldBeNil)
		So(d.URI, ShouldEqual, "www.google.com")
		So(d.SupplementaryFiles[0].Title, ShouldEqual, "helloworld")
		So(len(d.Downloads), ShouldEqual, 1)
		So(d.Downloads[0].File, ShouldEqual, "test.txt")
		So(d.Downloads[0].Size, ShouldEqual, strconv.Itoa(testFileSize))
	})

	Convey("test get dataset details when using a collection", t, func() {
		d, err := cli.GetDataset(ctx, testAccessToken, testCollectionID, testLang, "dataset")
		So(err, ShouldBeNil)
		So(d.URI, ShouldEqual, "www.google.com")
		So(d.SupplementaryFiles[0].Title, ShouldEqual, "helloworld")
		So(len(d.Downloads), ShouldEqual, 1)
		So(d.Downloads[0].File, ShouldEqual, "testCollection.txt")
		So(d.Downloads[0].Size, ShouldEqual, strconv.Itoa(testFileSizeCollection))
	})

	Convey("test getFileSize returns human readable filesize", t, func() {
		fs, err := cli.GetFileSize(ctx, testAccessToken, "", testLang, "filesize")
		So(err, ShouldBeNil)
		So(fs.Size, ShouldEqual, testFileSize)
	})

	Convey("test getFileSize returns human readable filesize when using a collection", t, func() {
		fs, err := cli.GetFileSize(ctx, testAccessToken, testCollectionID, testLang, "filesize")
		So(err, ShouldBeNil)
		So(fs.Size, ShouldEqual, testFileSizeCollection)
	})

	Convey("test getPageTitle returns a correctly formatted page title", t, func() {
		t, err := cli.GetPageTitle(ctx, testAccessToken, "", testLang, "pageTitle1")
		So(err, ShouldBeNil)
		So(t.Title, ShouldEqual, "baby-names")
		So(t.Edition, ShouldEqual, "2017")
		So(t.URI, ShouldEqual, "path/to/baby-names/2017")
	})

	Convey("test getPageTitle returns a correctly formatted page title when using a collection", t, func() {
		t, err := cli.GetPageTitle(ctx, testAccessToken, testCollectionID, testLang, "pageTitle1")
		So(err, ShouldBeNil)
		So(t.Title, ShouldEqual, "baby-names")
		So(t.Edition, ShouldEqual, "collection")
		So(t.URI, ShouldEqual, "path/to/baby-names/collection")
	})

	Convey("test GetPageDescription", t, func() {
		Convey("when not using a collection", func() {
			collectionId := ""
			Convey("it returns a page description", func() {
				d, err := cli.GetPageDescription(ctx, testAccessToken, collectionId, testLang, "pageDescription1")
				So(err, ShouldBeNil)
				So(d.URI, ShouldEqual, "path/to/page-description")
				So(d.Description.Title, ShouldEqual, "Page title")
				So(d.Description.Edition, ShouldEqual, "August 2015")
				So(d.Description.Summary, ShouldEqual, "This is the page summary")
				So(len(d.Description.Keywords), ShouldEqual, 2)
				So(d.Description.Keywords[0], ShouldEqual, "Economy")
				So(d.Description.Keywords[1], ShouldEqual, "Retail")
				So(d.Description.MetaDescription, ShouldEqual, "meta")
				So(d.Description.NationalStatistic, ShouldBeTrue)
				So(d.Description.LatestRelease, ShouldBeTrue)
				So(d.Description.ReleaseDate, ShouldEqual, "2015-09-14T23:00:00.000Z")
				So(d.Description.NextRelease, ShouldEqual, "13 October 2015")
				So(d.Description.Contact.Name, ShouldEqual, "Contact")
				So(d.Description.Contact.Email, ShouldEqual, "contact@ons.gov.uk")
				So(d.Description.Contact.Telephone, ShouldEqual, "+44 (0) 1633 456900")
			})
		})
		Convey("when using a collection", func() {
			collectionId := testCollectionID
			Convey("it returns a page description", func() {
				d, err := cli.GetPageDescription(ctx, testAccessToken, collectionId, testLang, "pageDescription1")
				So(err, ShouldBeNil)
				So(d.URI, ShouldEqual, "path/to/page-description/collection")
				So(d.Description.Title, ShouldEqual, "Page title")
				So(d.Description.Edition, ShouldEqual, "collection")
				So(d.Description.Summary, ShouldEqual, "This is the page summary")
				So(len(d.Description.Keywords), ShouldEqual, 2)
				So(d.Description.Keywords[0], ShouldEqual, "Economy")
				So(d.Description.Keywords[1], ShouldEqual, "Retail")
				So(d.Description.MetaDescription, ShouldEqual, "meta")
				So(d.Description.NationalStatistic, ShouldBeTrue)
				So(d.Description.LatestRelease, ShouldBeTrue)
				So(d.Description.ReleaseDate, ShouldEqual, "2015-09-14T23:00:00.000Z")
				So(d.Description.NextRelease, ShouldEqual, "13 October 2015")
				So(d.Description.Contact.Name, ShouldEqual, "Contact")
				So(d.Description.Contact.Email, ShouldEqual, "contact@ons.gov.uk")
				So(d.Description.Contact.Telephone, ShouldEqual, "+44 (0) 1633 456900")
			})
		})
		Convey("when the page type maps summary to _abstract", func() {
			collectionId := ""
			Convey("it maps to abstract", func() {
				d, err := cli.GetPageDescription(ctx, testAccessToken, collectionId, testLang, "pageDescription3")
				So(err, ShouldBeNil)
				So(d.Description.Title, ShouldEqual, "Another page title")
				So(d.Description.Abstract, ShouldEqual, "Page description is mapped from _abstract")
			})
		})
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
		Convey("when using a collection", func() {
			collectionId := testCollectionID
			Convey("returns the latest release of a bulletin", func() {
				b, err := cli.GetBulletin(ctx, testAccessToken, collectionId, testLang, "bulletin-latest-release")
				So(err, ShouldBeNil)
				So(b, ShouldNotBeEmpty)
				So(b.Type, ShouldEqual, "bulletin")
				So(b.URI, ShouldEqual, "/bulletin/2015-07-09")
				So(b.RelatedBulletins, ShouldNotBeEmpty)
				So(len(b.RelatedBulletins), ShouldEqual, 1)
				So(b.RelatedBulletins[0].URI, ShouldEqual, "pageTitle1")
				So(b.RelatedBulletins[0].Title, ShouldEqual, "baby-names: collection")
				So(b.Links, ShouldNotBeEmpty)
				So(len(b.Links), ShouldEqual, 2)
				So(b.Links[0].URI, ShouldEqual, "pageTitle1")
				So(b.Links[0].Title, ShouldEqual, "baby-names: collection")
				So(b.Links[1].URI, ShouldEqual, "pageTitle2")
				So(b.Links[1].Title, ShouldEqual, "page-title: c2021")
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
				So(b.Versions, ShouldNotBeEmpty)
				So(len(b.Versions), ShouldEqual, 1)
				So(b.Versions[0].ReleaseDate, ShouldEqual, "2021-10-19T10:43:34.507Z")
				So(b.Versions[0].Notice, ShouldEqual, "Notice")
				So(b.Versions[0].URI, ShouldEqual, "v1")
				So(b.Alerts, ShouldNotBeEmpty)
				So(len(b.Alerts), ShouldEqual, 1)
				So(b.Alerts[0].Markdown, ShouldEqual, "alert")
				So(b.Alerts[0].Date, ShouldEqual, "2021-09-30T07:10:46.230Z")
				So(b.Description, ShouldNotBeEmpty)
				So(b.Description.Title, ShouldEqual, "UK Environmental Accounts with collection")
				So(b.Description.Summary, ShouldEqual, "Measures the contribution of the environment to the economy")
				So(b.Description.MetaDescription, ShouldEqual, "Measures the contribution of the environment.")
				So(b.Description.NationalStatistic, ShouldBeTrue)
				So(b.Description.LatestRelease, ShouldBeTrue)
				So(b.LatestReleaseURI, ShouldBeBlank)
				So(b.Description.Edition, ShouldEqual, "2015")
				So(b.Description.ReleaseDate, ShouldEqual, "2015-07-08T23:00:00.000Z")
				So(b.Description.Contact, ShouldNotBeEmpty)
				So(b.Description.Contact.Email, ShouldEqual, "environment.accounts@ons.gsi.gov.uk")
				So(b.Description.Contact.Name, ShouldEqual, "Someone")
				So(b.Description.Contact.Telephone, ShouldEqual, "+44 (0)1633 455680")
			})

			Convey("returns a non-latest release of a bulletin", func() {
				b, err := cli.GetBulletin(ctx, testAccessToken, collectionId, testLang, "bulletin-not-latest-release")
				So(err, ShouldBeNil)
				So(b, ShouldNotBeEmpty)
				So(b.Type, ShouldEqual, "bulletin")
				So(b.URI, ShouldEqual, "/bulletin/2015-07-09")
				So(b.RelatedBulletins, ShouldNotBeEmpty)
				So(len(b.RelatedBulletins), ShouldEqual, 1)
				So(b.RelatedBulletins[0].URI, ShouldEqual, "pageTitle1")
				So(b.RelatedBulletins[0].Title, ShouldEqual, "baby-names: collection")
				So(b.Links, ShouldNotBeEmpty)
				So(len(b.Links), ShouldEqual, 2)
				So(b.Links[0].URI, ShouldEqual, "pageTitle1")
				So(b.Links[0].Title, ShouldEqual, "baby-names: collection")
				So(b.Links[1].URI, ShouldEqual, "pageTitle2")
				So(b.Links[1].Title, ShouldEqual, "page-title: c2021")
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
				So(b.Versions, ShouldNotBeEmpty)
				So(len(b.Versions), ShouldEqual, 1)
				So(b.Versions[0].ReleaseDate, ShouldEqual, "2021-10-19T10:43:34.507Z")
				So(b.Versions[0].Notice, ShouldEqual, "Notice")
				So(b.Versions[0].URI, ShouldEqual, "v1")
				So(b.Alerts, ShouldNotBeEmpty)
				So(len(b.Alerts), ShouldEqual, 1)
				So(b.Alerts[0].Markdown, ShouldEqual, "alert")
				So(b.Alerts[0].Date, ShouldEqual, "2021-09-30T07:10:46.230Z")
				So(b.Description, ShouldNotBeEmpty)
				So(b.Description.Title, ShouldEqual, "UK Environmental Accounts with collection")
				So(b.Description.Summary, ShouldEqual, "Measures the contribution of the environment to the economy")
				So(b.Description.MetaDescription, ShouldEqual, "Measures the contribution of the environment.")
				So(b.Description.NationalStatistic, ShouldBeTrue)
				So(b.Description.LatestRelease, ShouldBeFalse)
				So(b.LatestReleaseURI, ShouldEqual, "/bulletin/2021")
				So(b.Description.Edition, ShouldEqual, "2015")
				So(b.Description.ReleaseDate, ShouldEqual, "2015-07-08T23:00:00.000Z")
				So(b.Description.Contact, ShouldNotBeEmpty)
				So(b.Description.Contact.Email, ShouldEqual, "environment.accounts@ons.gsi.gov.uk")
				So(b.Description.Contact.Name, ShouldEqual, "Someone")
				So(b.Description.Contact.Telephone, ShouldEqual, "+44 (0)1633 455680")
			})
		})

		Convey("when not using a collection", func() {
			collectionId := ""
			Convey("returns the latest release of a bulletin", func() {
				b, err := cli.GetBulletin(ctx, testAccessToken, collectionId, testLang, "bulletin-latest-release")
				So(err, ShouldBeNil)
				So(b, ShouldNotBeEmpty)
				So(b.Type, ShouldEqual, "bulletin")
				So(b.URI, ShouldEqual, "/bulletin/2015-07-09")
				So(b.RelatedBulletins, ShouldNotBeEmpty)
				So(len(b.RelatedBulletins), ShouldEqual, 1)
				So(b.RelatedBulletins[0].URI, ShouldEqual, "pageTitle1")
				So(b.RelatedBulletins[0].Title, ShouldEqual, "baby-names: 2017")
				So(b.Links, ShouldNotBeEmpty)
				So(len(b.Links), ShouldEqual, 2)
				So(b.Links[0].URI, ShouldEqual, "pageTitle1")
				So(b.Links[0].Title, ShouldEqual, "baby-names: 2017")
				So(b.Links[1].URI, ShouldEqual, "pageTitle2")
				So(b.Links[1].Title, ShouldEqual, "page-title: 2021")
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
				So(b.Versions, ShouldNotBeEmpty)
				So(len(b.Versions), ShouldEqual, 1)
				So(b.Versions[0].ReleaseDate, ShouldEqual, "2021-10-19T10:43:34.507Z")
				So(b.Versions[0].Notice, ShouldEqual, "Notice")
				So(b.Versions[0].URI, ShouldEqual, "v1")
				So(b.Alerts, ShouldNotBeEmpty)
				So(len(b.Alerts), ShouldEqual, 1)
				So(b.Alerts[0].Markdown, ShouldEqual, "alert")
				So(b.Alerts[0].Date, ShouldEqual, "2021-09-30T07:10:46.230Z")
				So(b.Description, ShouldNotBeEmpty)
				So(b.Description.Title, ShouldEqual, "UK Environmental Accounts")
				So(b.Description.Summary, ShouldEqual, "Measures the contribution of the environment to the economy")
				So(b.Description.Title, ShouldEqual, "UK Environmental Accounts")
				So(b.Description.MetaDescription, ShouldEqual, "Measures the contribution of the environment.")
				So(b.Description.NationalStatistic, ShouldBeTrue)
				So(b.Description.LatestRelease, ShouldBeTrue)
				So(b.LatestReleaseURI, ShouldBeBlank)
				So(b.Description.Edition, ShouldEqual, "2015")
				So(b.Description.ReleaseDate, ShouldEqual, "2015-07-08T23:00:00.000Z")
				So(b.Description.Contact, ShouldNotBeEmpty)
				So(b.Description.Contact.Email, ShouldEqual, "environment.accounts@ons.gsi.gov.uk")
				So(b.Description.Contact.Name, ShouldEqual, "Someone")
				So(b.Description.Contact.Telephone, ShouldEqual, "+44 (0)1633 455680")
			})

			Convey("returns a non-latest release of a bulletin", func() {
				b, err := cli.GetBulletin(ctx, testAccessToken, collectionId, testLang, "bulletin-not-latest-release")
				So(err, ShouldBeNil)
				So(b, ShouldNotBeEmpty)
				So(b.Type, ShouldEqual, "bulletin")
				So(b.URI, ShouldEqual, "/bulletin/2015-07-09")
				So(b.RelatedBulletins, ShouldNotBeEmpty)
				So(len(b.RelatedBulletins), ShouldEqual, 1)
				So(b.RelatedBulletins[0].URI, ShouldEqual, "pageTitle1")
				So(b.RelatedBulletins[0].Title, ShouldEqual, "baby-names: 2017")
				So(b.Links, ShouldNotBeEmpty)
				So(len(b.Links), ShouldEqual, 2)
				So(b.Links[0].URI, ShouldEqual, "pageTitle1")
				So(b.Links[0].Title, ShouldEqual, "baby-names: 2017")
				So(b.Links[1].URI, ShouldEqual, "pageTitle2")
				So(b.Links[1].Title, ShouldEqual, "page-title: 2021")
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
				So(b.Versions, ShouldNotBeEmpty)
				So(len(b.Versions), ShouldEqual, 1)
				So(b.Versions[0].ReleaseDate, ShouldEqual, "2021-10-19T10:43:34.507Z")
				So(b.Versions[0].Notice, ShouldEqual, "Notice")
				So(b.Versions[0].URI, ShouldEqual, "v1")
				So(b.Alerts, ShouldNotBeEmpty)
				So(len(b.Alerts), ShouldEqual, 1)
				So(b.Alerts[0].Markdown, ShouldEqual, "alert")
				So(b.Alerts[0].Date, ShouldEqual, "2021-09-30T07:10:46.230Z")
				So(b.Description, ShouldNotBeEmpty)
				So(b.Description.Title, ShouldEqual, "UK Environmental Accounts")
				So(b.Description.Summary, ShouldEqual, "Measures the contribution of the environment to the economy")
				So(b.Description.MetaDescription, ShouldEqual, "Measures the contribution of the environment.")
				So(b.Description.NationalStatistic, ShouldBeTrue)
				So(b.Description.LatestRelease, ShouldBeFalse)
				So(b.LatestReleaseURI, ShouldEqual, "/bulletin/collection/2021")
				So(b.Description.Edition, ShouldEqual, "2015")
				So(b.Description.ReleaseDate, ShouldEqual, "2015-07-08T23:00:00.000Z")
				So(b.Description.Contact, ShouldNotBeEmpty)
				So(b.Description.Contact.Email, ShouldEqual, "environment.accounts@ons.gsi.gov.uk")
				So(b.Description.Contact.Name, ShouldEqual, "Someone")
				So(b.Description.Contact.Telephone, ShouldEqual, "+44 (0)1633 455680")
			})
		})

		Convey("returns an error if uri not found", func() {
			b, err := cli.GetBulletin(ctx, testAccessToken, "", testLang, "notFound")
			So(err, ShouldNotBeNil)
			So(b, ShouldResemble, Bulletin{})
		})
	})

	Convey("test GetRelease", t, func() {
		Convey("Given that we are not using a collection", func() {
			collectionId := ""
			Convey("When we call GetRelease with a valid clean path", func() {
				r, err := cli.GetRelease(ctx, testAccessToken, collectionId, testLang, "release")
				Convey("Then it returns the release", func() {
					So(err, ShouldBeNil)
					So(r, ShouldNotBeEmpty)
					So(r.URI, ShouldEqual, "/releases/indexofproductionukdecember2021timeseries")
					So(len(r.Markdown), ShouldEqual, 1)
					So(r.Markdown[0], ShouldEqual, "markdown")
					So(len(r.RelatedDocuments), ShouldEqual, 1)
					So(r.RelatedDocuments[0].URI, ShouldEqual, "pageDescription2")
					So(r.RelatedDocuments[0].Title, ShouldEqual, "UK Environmental Accounts: 2021")
					So(r.RelatedDocuments[0].Summary, ShouldEqual, "Measuring the contribution of the environment to the economy")
					So(len(r.RelatedDatasets), ShouldEqual, 1)
					So(r.RelatedDatasets[0].URI, ShouldEqual, "pageDescription1")
					So(r.RelatedDatasets[0].Title, ShouldEqual, "Page title: August 2015")
					So(r.RelatedDatasets[0].Summary, ShouldEqual, "This is the page summary")
					So(len(r.RelatedAPIDatasets), ShouldEqual, 2)
					So(r.RelatedAPIDatasets[0].URI, ShouldEqual, "cantabularDataset")
					So(r.RelatedAPIDatasets[0].Title, ShouldEqual, "Title for cantabularDataset")
					So(r.RelatedAPIDatasets[1].URI, ShouldEqual, "cmdDataset")
					So(r.RelatedAPIDatasets[1].Title, ShouldEqual, "Title for cmdDataset")
					So(len(r.RelatedMethodology), ShouldEqual, 1)
					So(r.RelatedMethodology[0].URI, ShouldEqual, "pageDescription1")
					So(r.RelatedMethodology[0].Title, ShouldEqual, "Page title: August 2015")
					So(r.RelatedMethodology[0].Summary, ShouldEqual, "This is the page summary")
					So(len(r.RelatedMethodologyArticle), ShouldEqual, 1)
					So(r.RelatedMethodologyArticle[0].URI, ShouldEqual, "pageDescription2")
					So(r.RelatedMethodologyArticle[0].Title, ShouldEqual, "UK Environmental Accounts: 2021")
					So(r.RelatedMethodologyArticle[0].Summary, ShouldEqual, "Measuring the contribution of the environment to the economy")
					So(len(r.Links), ShouldEqual, 3)
					So(r.Links[0].URI, ShouldEqual, "pageDescription1")
					So(r.Links[0].Title, ShouldEqual, "Page title: August 2015")
					So(r.Links[0].Summary, ShouldEqual, "This is the page summary")
					So(r.Links[1].URI, ShouldEqual, "pageDescription2")
					So(r.Links[1].Title, ShouldEqual, "UK Environmental Accounts: 2021")
					So(r.Links[1].Summary, ShouldEqual, "Measuring the contribution of the environment to the economy")
					So(r.Links[2].URI, ShouldEqual, "externalLinkURI")
					So(r.Links[2].Title, ShouldEqual, "This is a link to an external site")
					So(r.Links[2].Summary, ShouldEqual, "")
					So(len(r.DateChanges), ShouldEqual, 1)
					So(r.DateChanges[0].Date, ShouldEqual, "2021-08-15T11:12:05.592Z")
					So(r.DateChanges[0].ChangeNotice, ShouldEqual, "change notice")
					So(r.Description.Title, ShouldEqual, "Index of Production")
					So(r.Description.Summary, ShouldEqual, "Movements in the volume of production for the UK production industries")
					So(r.Description.NationalStatistic, ShouldBeTrue)
					So(r.Description.ReleaseDate, ShouldEqual, "2022-02-11T07:00:00.000Z")
					So(r.Description.NextRelease, ShouldEqual, "11 March 2022")
					So(r.Description.Contact.Name, ShouldEqual, "Contact name")
					So(r.Description.Contact.Email, ShouldEqual, "indexofproduction@ons.gov.uk")
					So(r.Description.Contact.Telephone, ShouldEqual, "+44 1633 456980")
					So(r.Description.Cancelled, ShouldBeTrue)
					So(len(r.Description.CancellationNotice), ShouldEqual, 1)
					So(r.Description.CancellationNotice[0], ShouldEqual, "notice")
					So(r.Description.Finalised, ShouldBeTrue)
					So(r.Description.Published, ShouldBeTrue)
					So(r.Description.ProvisionalDate, ShouldEqual, "Dec 22")
				})
			})

			Convey("When we call GetRelease with a dirty path", func() {
				r, err := cli.GetRelease(ctx, testAccessToken, collectionId, testLang, ".././/..///blah/./blah/../..//release")
				Convey("Then it cleans the path and returns the release", func() {
					So(err, ShouldBeNil)
					So(r, ShouldNotBeEmpty)
					So(r.URI, ShouldEqual, "/releases/indexofproductionukdecember2021timeseries")
					So(len(r.Markdown), ShouldEqual, 1)
					So(r.Markdown[0], ShouldEqual, "markdown")
					So(len(r.RelatedDocuments), ShouldEqual, 1)
					So(r.RelatedDocuments[0].URI, ShouldEqual, "pageDescription2")
					So(r.RelatedDocuments[0].Title, ShouldEqual, "UK Environmental Accounts: 2021")
					So(r.RelatedDocuments[0].Summary, ShouldEqual, "Measuring the contribution of the environment to the economy")
					So(len(r.RelatedDatasets), ShouldEqual, 1)
					So(r.RelatedDatasets[0].URI, ShouldEqual, "pageDescription1")
					So(r.RelatedDatasets[0].Title, ShouldEqual, "Page title: August 2015")
					So(r.RelatedDatasets[0].Summary, ShouldEqual, "This is the page summary")
					So(len(r.RelatedAPIDatasets), ShouldEqual, 2)
					So(r.RelatedAPIDatasets[0].URI, ShouldEqual, "cantabularDataset")
					So(r.RelatedAPIDatasets[0].Title, ShouldEqual, "Title for cantabularDataset")
					So(r.RelatedAPIDatasets[1].URI, ShouldEqual, "cmdDataset")
					So(r.RelatedAPIDatasets[1].Title, ShouldEqual, "Title for cmdDataset")
					So(len(r.RelatedMethodology), ShouldEqual, 1)
					So(r.RelatedMethodology[0].URI, ShouldEqual, "pageDescription1")
					So(r.RelatedMethodology[0].Title, ShouldEqual, "Page title: August 2015")
					So(r.RelatedMethodology[0].Summary, ShouldEqual, "This is the page summary")
					So(len(r.RelatedMethodologyArticle), ShouldEqual, 1)
					So(r.RelatedMethodologyArticle[0].URI, ShouldEqual, "pageDescription2")
					So(r.RelatedMethodologyArticle[0].Title, ShouldEqual, "UK Environmental Accounts: 2021")
					So(r.RelatedMethodologyArticle[0].Summary, ShouldEqual, "Measuring the contribution of the environment to the economy")
					So(len(r.Links), ShouldEqual, 3)
					So(r.Links[0].URI, ShouldEqual, "pageDescription1")
					So(r.Links[0].Title, ShouldEqual, "Page title: August 2015")
					So(r.Links[0].Summary, ShouldEqual, "This is the page summary")
					So(r.Links[1].URI, ShouldEqual, "pageDescription2")
					So(r.Links[1].Title, ShouldEqual, "UK Environmental Accounts: 2021")
					So(r.Links[1].Summary, ShouldEqual, "Measuring the contribution of the environment to the economy")
					So(r.Links[2].URI, ShouldEqual, "externalLinkURI")
					So(r.Links[2].Title, ShouldEqual, "This is a link to an external site")
					So(r.Links[2].Summary, ShouldEqual, "")
					So(len(r.DateChanges), ShouldEqual, 1)
					So(r.DateChanges[0].Date, ShouldEqual, "2021-08-15T11:12:05.592Z")
					So(r.DateChanges[0].ChangeNotice, ShouldEqual, "change notice")
					So(r.Description.Title, ShouldEqual, "Index of Production")
					So(r.Description.Summary, ShouldEqual, "Movements in the volume of production for the UK production industries")
					So(r.Description.NationalStatistic, ShouldBeTrue)
					So(r.Description.ReleaseDate, ShouldEqual, "2022-02-11T07:00:00.000Z")
					So(r.Description.NextRelease, ShouldEqual, "11 March 2022")
					So(r.Description.Contact.Name, ShouldEqual, "Contact name")
					So(r.Description.Contact.Email, ShouldEqual, "indexofproduction@ons.gov.uk")
					So(r.Description.Contact.Telephone, ShouldEqual, "+44 1633 456980")
					So(r.Description.Cancelled, ShouldBeTrue)
					So(len(r.Description.CancellationNotice), ShouldEqual, 1)
					So(r.Description.CancellationNotice[0], ShouldEqual, "notice")
					So(r.Description.Finalised, ShouldBeTrue)
					So(r.Description.Published, ShouldBeTrue)
					So(r.Description.ProvisionalDate, ShouldEqual, "Dec 22")
				})
			})
		})

		Convey("Given that we are using a collection", func() {
			collectionId := testCollectionID
			Convey("When we call GetRelease with a valid clean path", func() {
				r, err := cli.GetRelease(ctx, testAccessToken, collectionId, testLang, "/release")
				Convey("Then it returns the release", func() {
					So(err, ShouldBeNil)
					So(r, ShouldNotBeEmpty)
					So(r.URI, ShouldEqual, "/releases/collection")
					So(len(r.Markdown), ShouldEqual, 1)
					So(r.Markdown[0], ShouldEqual, "collection markdown")
					So(len(r.RelatedDocuments), ShouldEqual, 2)
					So(r.RelatedDocuments[0].URI, ShouldEqual, "pageDescription2")
					So(r.RelatedDocuments[0].Title, ShouldEqual, "Collection UK Environmental Accounts: 2021c")
					So(r.RelatedDocuments[0].Summary, ShouldEqual, "Measuring the contribution of the environment to the economy")
					So(r.RelatedDocuments[1].URI, ShouldEqual, "pageDescription3")
					So(r.RelatedDocuments[1].Title, ShouldEqual, "Another page title")
					So(r.RelatedDocuments[1].Summary, ShouldEqual, "Summary is from the _abstract field")
					So(len(r.RelatedDatasets), ShouldEqual, 1)
					So(r.RelatedDatasets[0].URI, ShouldEqual, "pageDescription1")
					So(r.RelatedDatasets[0].Title, ShouldEqual, "Page title: collection")
					So(r.RelatedDatasets[0].Summary, ShouldEqual, "This is the page summary")
					So(len(r.RelatedMethodology), ShouldEqual, 1)
					So(r.RelatedMethodology[0].URI, ShouldEqual, "pageDescription1")
					So(r.RelatedMethodology[0].Title, ShouldEqual, "Page title: collection")
					So(r.RelatedMethodology[0].Summary, ShouldEqual, "This is the page summary")
					So(len(r.RelatedMethodologyArticle), ShouldEqual, 1)
					So(r.RelatedMethodologyArticle[0].URI, ShouldEqual, "pageDescription2")
					So(r.RelatedMethodologyArticle[0].Title, ShouldEqual, "Collection UK Environmental Accounts: 2021c")
					So(r.RelatedMethodologyArticle[0].Summary, ShouldEqual, "Measuring the contribution of the environment to the economy")
					So(len(r.Links), ShouldEqual, 3)
					So(r.Links[0].URI, ShouldEqual, "pageDescription1")
					So(r.Links[0].Title, ShouldEqual, "Page title: collection")
					So(r.Links[0].Summary, ShouldEqual, "This is the page summary")
					So(r.Links[1].URI, ShouldEqual, "pageDescription2")
					So(r.Links[1].Title, ShouldEqual, "Collection UK Environmental Accounts: 2021c")
					So(r.Links[1].Summary, ShouldEqual, "Measuring the contribution of the environment to the economy")
					So(r.Links[2].URI, ShouldEqual, "externalLinkURI")
					So(r.Links[2].Title, ShouldEqual, "This is a link to an external site")
					So(r.Links[2].Summary, ShouldEqual, "")
					So(len(r.DateChanges), ShouldEqual, 1)
					So(r.DateChanges[0].Date, ShouldEqual, "2021-08-15T11:12:05.592Z")
					So(r.DateChanges[0].ChangeNotice, ShouldEqual, "change notice")
					So(r.Description.Title, ShouldEqual, "Index of Production")
					So(r.Description.Summary, ShouldEqual, "Movements in the volume of production for the UK production industries")
					So(r.Description.NationalStatistic, ShouldBeTrue)
					So(r.Description.ReleaseDate, ShouldEqual, "2022-02-11T07:00:00.000Z")
					So(r.Description.NextRelease, ShouldEqual, "11 March 2022")
					So(r.Description.Contact.Name, ShouldEqual, "Contact name")
					So(r.Description.Contact.Email, ShouldEqual, "indexofproduction@ons.gov.uk")
					So(r.Description.Contact.Telephone, ShouldEqual, "+44 1633 456980")
					So(r.Description.Cancelled, ShouldBeTrue)
					So(len(r.Description.CancellationNotice), ShouldEqual, 1)
					So(r.Description.CancellationNotice[0], ShouldEqual, "notice")
					So(r.Description.Finalised, ShouldBeTrue)
					So(r.Description.Published, ShouldBeTrue)
					So(r.Description.ProvisionalDate, ShouldEqual, "Dec 22")
				})
			})

			Convey("When we call GetRelease with a dirty path", func() {
				r, err := cli.GetRelease(ctx, testAccessToken, collectionId, testLang, "/../././///blah/..///release")
				Convey("Then it returns the release", func() {
					So(err, ShouldBeNil)
					So(r, ShouldNotBeEmpty)
					So(r.URI, ShouldEqual, "/releases/collection")
					So(len(r.Markdown), ShouldEqual, 1)
					So(r.Markdown[0], ShouldEqual, "collection markdown")
					So(len(r.RelatedDocuments), ShouldEqual, 2)
					So(r.RelatedDocuments[0].URI, ShouldEqual, "pageDescription2")
					So(r.RelatedDocuments[0].Title, ShouldEqual, "Collection UK Environmental Accounts: 2021c")
					So(r.RelatedDocuments[0].Summary, ShouldEqual, "Measuring the contribution of the environment to the economy")
					So(r.RelatedDocuments[1].URI, ShouldEqual, "pageDescription3")
					So(r.RelatedDocuments[1].Title, ShouldEqual, "Another page title")
					So(r.RelatedDocuments[1].Summary, ShouldEqual, "Summary is from the _abstract field")
					So(len(r.RelatedDatasets), ShouldEqual, 1)
					So(r.RelatedDatasets[0].URI, ShouldEqual, "pageDescription1")
					So(r.RelatedDatasets[0].Title, ShouldEqual, "Page title: collection")
					So(r.RelatedDatasets[0].Summary, ShouldEqual, "This is the page summary")
					So(len(r.RelatedMethodology), ShouldEqual, 1)
					So(r.RelatedMethodology[0].URI, ShouldEqual, "pageDescription1")
					So(r.RelatedMethodology[0].Title, ShouldEqual, "Page title: collection")
					So(r.RelatedMethodology[0].Summary, ShouldEqual, "This is the page summary")
					So(len(r.RelatedMethodologyArticle), ShouldEqual, 1)
					So(r.RelatedMethodologyArticle[0].URI, ShouldEqual, "pageDescription2")
					So(r.RelatedMethodologyArticle[0].Title, ShouldEqual, "Collection UK Environmental Accounts: 2021c")
					So(r.RelatedMethodologyArticle[0].Summary, ShouldEqual, "Measuring the contribution of the environment to the economy")
					So(len(r.Links), ShouldEqual, 3)
					So(r.Links[0].URI, ShouldEqual, "pageDescription1")
					So(r.Links[0].Title, ShouldEqual, "Page title: collection")
					So(r.Links[0].Summary, ShouldEqual, "This is the page summary")
					So(r.Links[1].URI, ShouldEqual, "pageDescription2")
					So(r.Links[1].Title, ShouldEqual, "Collection UK Environmental Accounts: 2021c")
					So(r.Links[1].Summary, ShouldEqual, "Measuring the contribution of the environment to the economy")
					So(r.Links[2].URI, ShouldEqual, "externalLinkURI")
					So(r.Links[2].Title, ShouldEqual, "This is a link to an external site")
					So(r.Links[2].Summary, ShouldEqual, "")
					So(len(r.DateChanges), ShouldEqual, 1)
					So(r.DateChanges[0].Date, ShouldEqual, "2021-08-15T11:12:05.592Z")
					So(r.DateChanges[0].ChangeNotice, ShouldEqual, "change notice")
					So(r.Description.Title, ShouldEqual, "Index of Production")
					So(r.Description.Summary, ShouldEqual, "Movements in the volume of production for the UK production industries")
					So(r.Description.NationalStatistic, ShouldBeTrue)
					So(r.Description.ReleaseDate, ShouldEqual, "2022-02-11T07:00:00.000Z")
					So(r.Description.NextRelease, ShouldEqual, "11 March 2022")
					So(r.Description.Contact.Name, ShouldEqual, "Contact name")
					So(r.Description.Contact.Email, ShouldEqual, "indexofproduction@ons.gov.uk")
					So(r.Description.Contact.Telephone, ShouldEqual, "+44 1633 456980")
					So(r.Description.Cancelled, ShouldBeTrue)
					So(len(r.Description.CancellationNotice), ShouldEqual, 1)
					So(r.Description.CancellationNotice[0], ShouldEqual, "notice")
					So(r.Description.Finalised, ShouldBeTrue)
					So(r.Description.Published, ShouldBeTrue)
					So(r.Description.ProvisionalDate, ShouldEqual, "Dec 22")
				})
			})
		})
	})

}

func TestClient_HealthChecker(t *testing.T) {
	t.Parallel()
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
	t.Parallel()
	ctx := context.Background()
	path := "/publisheddata"
	testURIString := "/employmentandlabourmarket/peopleinwork/workplacedisputesandworkingconditions/datasets/labourdisputesbysectorlabd02"
	documentContent := []byte("{byte slice returned}")
	body := httpmocks.NewReadCloserMock(documentContent, nil)

	Convey("given a 200 response", t, func() {
		response := httpmocks.NewResponseMock(body, http.StatusOK)
		httpClient := newMockHTTPClient(response, nil)
		zebedeeClient := newZebedeeClient(httpClient)

		Convey("when zebedeeClient.GetPublishedData is called", func() {
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

func TestClient_PublishedIndexEndpoint(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	path := "/publishedindex"

	Convey("given a 200 response", t, func() {
		documentContent, err := os.ReadFile("./response_mocks/publishedindex.json")
		So(err, ShouldBeNil)
		body := httpmocks.NewReadCloserMock(documentContent, nil)
		response := httpmocks.NewResponseMock(body, http.StatusOK)

		httpClient := newMockHTTPClient(response, nil)
		zebedeeClient := newZebedeeClient(httpClient)

		Convey("when zebedeeClient.GetPublishedIndex is called", func() {
			m, err := zebedeeClient.GetPublishedIndex(ctx, nil)

			Convey("then the expected content is returned", func() {
				So(err, ShouldBeNil)
				So(m.Count, ShouldEqual, 3)
				So(m.Items, ShouldNotBeEmpty)
				So(m.Items, ShouldHaveLength, 3)
				mItem := m.Items[0]
				So(mItem.URI, ShouldResemble, "/releases")
				So(m.Limit, ShouldEqual, 3)
				So(m.Offset, ShouldEqual, 0)
				So(m.TotalCount, ShouldEqual, 3)
			})

			Convey("and client.Do should be called once with the expected parameters", func() {
				doCalls := httpClient.DoCalls()
				So(doCalls, ShouldHaveLength, 1)
				So(doCalls[0].Req.URL.Path, ShouldEqual, path)
			})
		})

		Convey("when zebedeeClient.GetPublishedIndex is called with paging params", func() {
			limit := 9
			m, err := zebedeeClient.GetPublishedIndex(ctx, &PublishedIndexRequestParams{
				Limit:  &limit,
				Offset: 5,
			})

			Convey("then the expected content is returned", func() {
				So(err, ShouldBeNil)
				So(m.Count, ShouldEqual, 3)
				So(m.Items, ShouldNotBeEmpty)
				So(m.Items, ShouldHaveLength, 3)
				mItem := m.Items[0]
				So(mItem.URI, ShouldResemble, "/releases")
				So(m.Limit, ShouldEqual, 3)
				So(m.Offset, ShouldEqual, 0)
				So(m.TotalCount, ShouldEqual, 3)
			})

			Convey("and client.Do should be called once with the expected parameters", func() {
				doCalls := httpClient.DoCalls()
				So(doCalls, ShouldHaveLength, 1)
				So(doCalls[0].Req.URL.Path, ShouldEqual, path)
				req := doCalls[0].Req
				So(req.FormValue("offset"), ShouldEqual, "5")
				So(req.FormValue("limit"), ShouldEqual, "9")
			})
		})
	})

	Convey("given a 500 response", t, func() {
		documentContent := []byte("{byte slice returned}")
		body := httpmocks.NewReadCloserMock(documentContent, nil)
		response := httpmocks.NewResponseMock(body, http.StatusInternalServerError)
		httpClient := newMockHTTPClient(response, nil)
		zebedeeClient := newZebedeeClient(httpClient)

		Convey("when zebedeeClient.GetPublishedIndex is called", func() {
			_, err := zebedeeClient.GetPublishedIndex(ctx, nil)

			Convey("then the expected error is returned", func() {
				So(err, ShouldNotBeNil)
				So(err, ShouldHaveSameTypeAs, ErrInvalidZebedeeResponse{})
				So(err.Error(), ShouldEqual, "invalid response from zebedee: 500, path: /publishedindex")
			})

			Convey("and client.Do should be called once with the expected parameters", func() {
				doCalls := httpClient.DoCalls()
				So(doCalls, ShouldHaveLength, 1)
				So(doCalls[0].Req.URL.Path, ShouldEqual, path)
			})
		})
	})
}

func MockZebedeeDatasetHandler(mockDataset Dataset, expectedFileSize int, fileNotExist bool) http.HandlerFunc {
	mockFileSize := FileSize{Size: expectedFileSize}

	return func(w http.ResponseWriter, r *http.Request) {
		je := json.NewEncoder(w)

		switch r.URL.Path {
		case "/data":
			je.Encode(mockDataset)
		case "/filesize":
			if fileNotExist {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			je.Encode(mockFileSize)
		}
	}
}

func TestGetDataset(t *testing.T) {
	Convey("Client.GetDataset called and file exists within Zebedee", t, func() {
		Convey("result contains valid downloads", func() {
			expectedDownloadFile := "filename.csv"
			expectedSize := 100
			mockDownload := Download{File: expectedDownloadFile}
			mockDataset := Dataset{
				Downloads: []Download{mockDownload},
			}

			mockDatasetServer := httptest.NewServer(MockZebedeeDatasetHandler(mockDataset, expectedSize, false))
			defer mockDatasetServer.Close()

			client := New(mockDatasetServer.URL)
			result, err := client.GetDataset(context.Background(), "", "", "", "")
			So(err, ShouldBeNil)
			So(len(result.Downloads), ShouldEqual, 1)

			firstDownloadResult := result.Downloads[0]
			actualFileSize, _ := strconv.Atoi(firstDownloadResult.Size)
			actualFilename := firstDownloadResult.File

			So(actualFilename, ShouldEqual, expectedDownloadFile)
			So(actualFileSize, ShouldEqual, expectedSize)
		})

		Convey("result contains valid supplementary files", func() {
			expectedSupplementaryTitle := "title"
			expectedSupplementaryFile := "file.csv"
			expectedSize := 100

			mockSupplementaryFile := SupplementaryFile{
				Title: expectedSupplementaryTitle,
				File:  expectedSupplementaryFile,
			}

			mockDataset := Dataset{
				SupplementaryFiles: []SupplementaryFile{mockSupplementaryFile},
			}

			mockDatasetServer := httptest.NewServer(MockZebedeeDatasetHandler(mockDataset, expectedSize, false))
			defer mockDatasetServer.Close()

			client := New(mockDatasetServer.URL)
			result, err := client.GetDataset(context.Background(), "", "", "", "")

			So(err, ShouldBeNil)
			So(len(result.SupplementaryFiles), ShouldEqual, 1)

			firstSupplementaryFileResult := result.SupplementaryFiles[0]
			actualSupplementaryFileSize, _ := strconv.Atoi(firstSupplementaryFileResult.Size)
			actualSupplementaryFilename := firstSupplementaryFileResult.File

			So(actualSupplementaryFilename, ShouldEqual, expectedSupplementaryFile)
			So(actualSupplementaryFileSize, ShouldEqual, expectedSize)
		})
	})

	Convey("Client.GetDataset called and file exists in files API", t, func() {
		Convey("result contains a download without size", func() {
			expectedDownloadFile := "filename.csv"
			expectedSize := 100
			expectedVersion := "2"
			mockDownload := Download{
				URI:     expectedDownloadFile,
				Version: expectedVersion,
			}
			mockDataset := Dataset{
				Downloads: []Download{mockDownload},
			}

			mockZebedeeServer := httptest.NewServer(MockZebedeeDatasetHandler(mockDataset, expectedSize, true))
			defer mockZebedeeServer.Close()

			client := New(mockZebedeeServer.URL)
			result, err := client.GetDataset(context.Background(), "", "", "", "")
			So(err, ShouldBeNil)
			So(len(result.Downloads), ShouldEqual, 1)

			firstDownloadResult := result.Downloads[0]
			actualFileSize, _ := strconv.Atoi(firstDownloadResult.Size)
			actualFilename := firstDownloadResult.URI

			So(actualFilename, ShouldEqual, expectedDownloadFile)
			So(actualFileSize, ShouldEqual, 0)
		})

		Convey("result contains a supplementary file without size", func() {
			expectedSupplementaryFile := "filename.csv"
			expectedSize := 100
			expectedVersion := "2"
			mockSupplementaryFile := SupplementaryFile{
				URI:     expectedSupplementaryFile,
				Version: expectedVersion,
			}
			mockDataset := Dataset{
				SupplementaryFiles: []SupplementaryFile{mockSupplementaryFile},
			}

			mockZebedeeServer := httptest.NewServer(MockZebedeeDatasetHandler(mockDataset, expectedSize, true))
			defer mockZebedeeServer.Close()

			client := New(mockZebedeeServer.URL)
			result, err := client.GetDataset(context.Background(), "", "", "", "")
			So(err, ShouldBeNil)
			So(len(result.SupplementaryFiles), ShouldEqual, 1)

			firstSupplementaryFileResult := result.SupplementaryFiles[0]
			actualFileSize, _ := strconv.Atoi(firstSupplementaryFileResult.Size)
			actualFilename := firstSupplementaryFileResult.URI

			So(actualFilename, ShouldEqual, expectedSupplementaryFile)
			So(actualFileSize, ShouldEqual, 0)
		})
	})
}
