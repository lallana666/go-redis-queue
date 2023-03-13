package xiaotao

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/golang/glog"
)

// ValuateResult is the return value of Valuate.
type ValuateResult struct {
	Price       float64
	CommunityID string
}

// Valuate 估值
func Valuate(address string, cityCode string, area float64, kind string) (*ValuateResult, error) {
	glog.Infof("xiaotao: Valuate(%q, %q, %d, %q", address, cityCode, area, kind)

	if kind != "住宅" {
		return nil, errors.New("ethan: 小淘现在只能帮您计算「住宅」的估值")
	}

	rawres, err := Get(
		"/general/price/getEnquiryPrice/v3?" +
			"cityCode=" + cityCode +
			"&address=" + url.QueryEscape(address) +
			"&houseType=" + url.QueryEscape(kind) +
			"&buildingArea=" + strconv.FormatFloat(area, 'f', -1, 64))
	if err != nil {
		return nil, err
	}
	defer rawres.Body.Close()

	if rawres.StatusCode != http.StatusOK {
		buf, err := ioutil.ReadAll(rawres.Body)
		if err != nil {
			return nil, err
		}
		glog.Errorf("xiaotao: downstream status code %d (should be 200), body: %s", rawres.StatusCode, string(buf))
		return nil, fmt.Errorf("downstream status code: %d (should be 200)", rawres.StatusCode)
	}

	ct := rawres.Header.Get("Content-Type")
	if !strings.Contains(ct, "application/json") {
		buf, err := ioutil.ReadAll(rawres.Body)
		if err != nil {
			return nil, err
		}
		glog.Errorf("xiaotao: downstream content type %s (should be application/json), body: %s", ct, string(buf))
		return nil, fmt.Errorf("downstream content type: %s (should be application/json)", ct)
	}

	var res GetEnquiryPriceResponse
	if err = json.NewDecoder(rawres.Body).Decode(&res); err != nil {
		return nil, err
	}

	if res.ErrorCode != "OK" {
		glog.Errorf("xiaotao: downstream error code %q (should be 'OK'), response: %v", res.ErrorCode, res)
		return nil, fmt.Errorf("downstream error code: %q (should be 'OK')", res.ErrorCode)
	}

	if res.Data[0].StatusCode != "ENQUIRY-SUCCESS" {
		glog.Errorf("xiaotao: downstream status code %q (should be 'ENQUIRY-SUCCESS'), response: %v", res.Data[0].StatusCode, res)
		return nil, fmt.Errorf("downstream status code: %q (should be 'ENQUIRY-SUCCESS')", res.Data[0].StatusCode)
	}

	return &ValuateResult{
		Price:       res.Data[0].TotalPrice,
		CommunityID: res.Data[0].CommunityID,
	}, nil
}

// AcquireNeighboringPriceMap 获取抵押物周边房价地图
func AcquireNeighboringPriceMap(cityCode, communityID string) (string, error) {
	glog.Infof("xiaotao: AcquireNeighboringPriceMap(%q, %q)", cityCode, communityID)

	rawres, err := Get(
		"/general/price/getCommunityPeripheryAvgeragePrice/v3?" +
			"cityCode=" + cityCode +
			"&communityID=" + url.QueryEscape(communityID) +
			"&distance=500")
	if err != nil {
		return "", err
	}
	defer rawres.Body.Close()

	if rawres.StatusCode != http.StatusOK {
		buf, err := ioutil.ReadAll(rawres.Body)
		if err != nil {
			return "", err
		}
		glog.Errorf("xiaotao: downstream status code %d (should be 200), body: %s", rawres.StatusCode, string(buf))
		return "", fmt.Errorf("downstream status code: %d (should be 200)", rawres.StatusCode)
	}

	ct := rawres.Header.Get("Content-Type")
	if !strings.Contains(ct, "application/json") {
		buf, err := ioutil.ReadAll(rawres.Body)
		if err != nil {
			return "", err
		}
		glog.Errorf("xiaotao: downstream content type %s (should be application/json), body: %s", rawres.StatusCode, string(buf))
		return "", fmt.Errorf("downstream content type: %s (should be application/json)", ct)
	}

	var res GetCommunityPeripheryAvgeragePriceResponse
	if err = json.NewDecoder(rawres.Body).Decode(&res); err != nil {
		return "", err
	}

	if res.ErrorCode != "OK" {
		glog.Errorf("xiaotao: downstream error code %q (should be 'OK'), response: %v", res.ErrorCode, res)
		return "", fmt.Errorf("downstream error code: %q (should be 'OK')", res.ErrorCode)
	}

	buf := bytes.NewBufferString("")
	if err = json.NewEncoder(buf).Encode(res.Data[0].List); err != nil {
		return "", err
	}

	return buf.String(), nil
}

// AcquireCommunityInfo 获取抵押物所在小区的信息
func AcquireCommunityInfo(cityCode, communityID string) (string, error) {
	glog.Infof("xiaotao: AcquireCommunityInfo(%q, %q)", cityCode, communityID)

	rawres, err := Get(
		"/general/baseinfo/getCommunityBasicInfo/v3?" +
			"cityCode=" + cityCode +
			"&communityID=" + url.QueryEscape(communityID))
	if err != nil {
		return "", err
	}
	defer rawres.Body.Close()

	if rawres.StatusCode != http.StatusOK {
		buf, err := ioutil.ReadAll(rawres.Body)
		if err != nil {
			return "", err
		}
		glog.Errorf("xiaotao: downstream status code %d (should be 200), body: %s", rawres.StatusCode, string(buf))
		return "", fmt.Errorf("downstream status code: %d (should be 200)", rawres.StatusCode)
	}

	ct := rawres.Header.Get("Content-Type")
	if !strings.Contains(ct, "application/json") {
		buf, err := ioutil.ReadAll(rawres.Body)
		if err != nil {
			return "", err
		}
		glog.Errorf("xiaotao: downstream content type %s (should be application/json), body: %s", rawres.StatusCode, string(buf))
		return "", fmt.Errorf("downstream content type: %s (should be application/json)", ct)
	}

	buf, err := ioutil.ReadAll(rawres.Body)
	if err != nil {
		return "", err
	}

	return string(buf), nil
}

// AcquireResidentialFacilities 获取抵押物周边配套设施
func AcquireResidentialFacilities(cityCode, communityID string) (string, error) {
	glog.Infof("xiaotao: AcquireResidentialFacilities(%q, %q)", cityCode, communityID)

	rawres, err := Get(
		"/general/baseinfo/getCommunitySupporting/v3?" +
			"cityCode=" + cityCode +
			"&communityID=" + url.QueryEscape(communityID) +
			"&distance=2000")
	if err != nil {
		return "", err
	}
	defer rawres.Body.Close()

	if rawres.StatusCode != http.StatusOK {
		buf, err := ioutil.ReadAll(rawres.Body)
		if err != nil {
			return "", err
		}
		glog.Errorf("xiaotao: downstream status code %d (should be 200), body: %s", rawres.StatusCode, string(buf))
		return "", fmt.Errorf("downstream status code: %d (should be 200)", rawres.StatusCode)
	}

	ct := rawres.Header.Get("Content-Type")
	if !strings.Contains(ct, "application/json") {
		buf, err := ioutil.ReadAll(rawres.Body)
		if err != nil {
			return "", err
		}
		glog.Errorf("xiaotao: downstream content type %s (should be application/json), body: %s", rawres.StatusCode, string(buf))
		return "", fmt.Errorf("downstream content type: %s (should be application/json)", ct)
	}

	buf, err := ioutil.ReadAll(rawres.Body)
	if err != nil {
		return "", err
	}

	return string(buf), nil
}

// AcquirePawnCommunitySecondHandHousingTransactions 获取抵押物所在小区的二手房成交案例
func AcquirePawnCommunitySecondHandHousingTransactions(cityCode, communityID string) (string, error) {
	glog.Infof("xiaotao: AcquirePawnCommunitySecondHandHousingTransactions(%q, %q)", cityCode, communityID)

	rawres, err := Get(
		"/general/collateral/getSecondHandHousingDistrictTransactionCase/v3?" +
			"cityCode=" + cityCode +
			"&communityID=" + url.QueryEscape(communityID))
	if err != nil {
		return "", err
	}
	defer rawres.Body.Close()

	if rawres.StatusCode != http.StatusOK {
		buf, err := ioutil.ReadAll(rawres.Body)
		if err != nil {
			return "", err
		}
		glog.Errorf("xiaotao: downstream status code %d (should be 200), body: %s", rawres.StatusCode, string(buf))
		return "", fmt.Errorf("downstream status code: %d (should be 200)", rawres.StatusCode)
	}

	ct := rawres.Header.Get("Content-Type")
	if !strings.Contains(ct, "application/json") {
		buf, err := ioutil.ReadAll(rawres.Body)
		if err != nil {
			return "", err
		}
		glog.Errorf("xiaotao: downstream content type %s (should be application/json), body: %s", rawres.StatusCode, string(buf))
		return "", fmt.Errorf("downstream content type: %s (should be application/json)", ct)
	}

	buf, err := ioutil.ReadAll(rawres.Body)
	if err != nil {
		return "", err
	}

	return string(buf), nil
}

// AcquirePawnAveragePriceTrend 获取抵押物所在小区、行政区、城市的均值走势
func AcquirePawnAveragePriceTrend(cityCode, communityID string) (string, error) {
	glog.Infof("xiaotao: AcquirePawnAveragePriceTrend(%q, %q)", cityCode, communityID)

	rawres, err := Get(
		"/general/price/getAveragePriceOfUrbanDistrictTrend/v3?" +
			"cityCode=" + cityCode +
			"&communityID=" + url.QueryEscape(communityID) +
			"&timeSpan=12")
	if err != nil {
		return "", err
	}
	defer rawres.Body.Close()

	if rawres.StatusCode != http.StatusOK {
		buf, err := ioutil.ReadAll(rawres.Body)
		if err != nil {
			return "", err
		}
		glog.Errorf("xiaotao: downstream status code %d (should be 200), body: %s", rawres.StatusCode, string(buf))
		return "", fmt.Errorf("downstream status code: %d (should be 200)", rawres.StatusCode)
	}

	ct := rawres.Header.Get("Content-Type")
	if !strings.Contains(ct, "application/json") {
		buf, err := ioutil.ReadAll(rawres.Body)
		if err != nil {
			return "", err
		}
		glog.Errorf("xiaotao: downstream content type %s (should be application/json), body: %s", rawres.StatusCode, string(buf))
		return "", fmt.Errorf("downstream content type: %s (should be application/json)", ct)
	}

	buf, err := ioutil.ReadAll(rawres.Body)
	if err != nil {
		return "", err
	}

	return string(buf), nil
}

// AcquireCommunityRating 获取小区评级
func AcquireCommunityRating(cityCode, communityID string) (*GetCommunityRatingResponseData, error) {
	glog.Infof("xiaotao: AcquireCommunityRating(%q, %q)", cityCode, communityID)

	rawres, err := Get(
		"/general/grade/getCommunityRating/v3?" +
			"cityCode=" + cityCode +
			"&communityID=" + url.QueryEscape(communityID))
	if err != nil {
		return nil, err
	}
	defer rawres.Body.Close()

	if rawres.StatusCode != http.StatusOK {
		buf, err := ioutil.ReadAll(rawres.Body)
		if err != nil {
			return nil, err
		}
		glog.Errorf("xiaotao: downstream status code %d (should be 200), body: %s", rawres.StatusCode, string(buf))
		return nil, fmt.Errorf("downstream status code: %d (should be 200)", rawres.StatusCode)
	}

	ct := rawres.Header.Get("Content-Type")
	if !strings.Contains(ct, "application/json") {
		buf, err := ioutil.ReadAll(rawres.Body)
		if err != nil {
			return nil, err
		}
		glog.Errorf("xiaotao: downstream content type %s (should be application/json), body: %s", rawres.StatusCode, string(buf))
		return nil, fmt.Errorf("downstream content type: %s (should be application/json)", ct)
	}

	var res GetCommunityRatingResponse
	if err = json.NewDecoder(rawres.Body).Decode(&res); err != nil {
		return nil, err
	}

	if res.ErrorCode != "OK" {
		glog.Errorf("xiaotao: downstream error code %q (should be 'OK'), response: %v", res.ErrorCode, res)
		return nil, fmt.Errorf("downstream error code: %q (should be 'OK')", res.ErrorCode)
	}

	return &res.Data[0], nil
}
