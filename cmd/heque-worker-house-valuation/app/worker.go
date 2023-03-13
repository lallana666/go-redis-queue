package app

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"denggotech.cn/heque/heque/client"
	"denggotech.cn/heque/heque/cmd/heque-worker-house-valuation/app/types"
	utilxiaotao "denggotech.cn/heque/heque/cmd/heque-worker-house-valuation/xiaotao"
	utilflag "denggotech.cn/heque/heque/util/flag"
)

// 云房估值消费实体类
type jobArgs struct {
	//  pkg id
	PkgID string `json:"PkgID"`
	//  财产线索 ID
	PropertyID string `json:"PropertyID"`
	//  住宅地址
	Address string `json:"Address"`
	//  住宅面积
	Area string `json:"Area"`
	//  城市
	CityCode string `json:"CityCode"`
	//  住宅类型
	Type string `json:"Type"`
	//  估值token
	Token string `json:"Token"`
	//  批次id
	Batch string `json:"Batch"`
}

func NewWorkerCommand() *cobra.Command {
	s := NewWorkerOptions()

	cmd := &cobra.Command{
		Use:  "heque-worker-house-valuation",
		Long: `This worker valuates house.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			utilflag.PrintFlags(cmd.Flags())

			if err := s.Validate(); err != nil {
				return err
			}

			return Run(s)
		},
	}

	s.AddFlags(cmd.Flags())

	return cmd
}

var isMock string

func initIsMock(ismock string) error {
	isMock = ismock

	// no error
	return nil
}

// Run runs the specified worker. This should never exit.
func Run(w *WorkerOptions) error {
	// Initialize the xiaotao code
	err := utilxiaotao.Init(w.YunfangKeyID, []byte(w.YunfangAccessKey), w.YunfangDomain)
	if err != nil {
		return err
	}

	// Initialize the fahai-api ismock
	err = initIsMock(w.IsMock)
	if err != nil {
		return err
	}

	// Initialize heque client
	cli, err := client.New(client.Config{
		Endpoints: []string{w.RedisAddress},
	})
	if err != nil {
		return err
	}

	for {
		job, err := cli.Dequeue(w.QueueName)
		if err != nil {
			fmt.Println(err)
			continue
		}
		consumeOneJob(job, w.DebtdbAddress, cli)

		// 以后云房估值调用后可以去除sleep，现在模拟进度条间隔1秒
		time.Sleep(1 * time.Second)
	}

	return nil
}

func consumeOneJob(j *client.Job, debtdbAdress string, cli *client.Client) error {
	var jobArgs jobArgs

	err := json.Unmarshal([]byte(j.Spec.Payload), &jobArgs)
	if err != nil {
		fmt.Println(err)
		return err
	}

	// 估值
	area, err := strconv.ParseFloat(jobArgs.Area, 64)
	if err != nil {
		cli.MarkAsFailed(j)
		fmt.Println(err)
		return err
	}
	valuationAmount, err := valuateHouse(jobArgs.Address, area, jobArgs.CityCode, jobArgs.Type)
	if err != nil {
		cli.MarkAsFailed(j)
		fmt.Println(err)
		return err
	}

	// go graphql
	url := debtdbAdress

	// graphql 估值写回debtdb
	payloadStrUpdateVal := fmt.Sprintf("{\"query\":\"mutation ($input: UpdateHouseValuationInput!) {updateHouseValuation (input: $input) {valuation}}\",\"variables\":{\"input\":{\"houseId\":\"%s\",\"valuationAmount\":\"%f\"}}}", jobArgs.PropertyID, valuationAmount)

	updateValuationResponse, err := httpGraphqlValuationMutation(&jobArgs, payloadStrUpdateVal, url)
	if err != nil {
		cli.MarkAsFailed(j)
		return err
	}
	if updateValuationResponse.Errors != nil {
		cli.MarkAsFailed(j)
		fmt.Println(updateValuationResponse.Errors)
		return errors.New("更新估值报错")
	}

	cli.MarkAsDone(j)
	fmt.Println("房屋估值结束......jobId:" + j.ID)
	return nil
}

func httpGraphqlValuationMutation(j *jobArgs, payloadStr string, url string) (*types.ValuationResponse, error) {
	payload := strings.NewReader(payloadStr)
	client := &http.Client{}
	req, err := http.NewRequest("POST", url, payload)

	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer "+j.Token)

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	if res.StatusCode != http.StatusOK {
		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return nil, err
		}
		fmt.Println(string(body))
		return nil, errors.New("查询速效价报错")
	}

	var vresp types.ValuationResponse
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	fmt.Println(string(body))
	err = json.Unmarshal([]byte(string(body)), &vresp)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	return &vresp, err
}

func httpGraphqlValuationQuery(j *jobArgs, payloadStr string, url string) (*types.GetValuationResponse, error) {
	payload := strings.NewReader(payloadStr)
	client := &http.Client{}
	req, err := http.NewRequest("POST", url, payload)

	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer "+j.Token)

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	if res.StatusCode != http.StatusOK {
		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return nil, err
		}
		fmt.Println(string(body))
		return nil, errors.New("查询速效价报错")
	}

	var vresp types.GetValuationResponse
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	fmt.Println(string(body))
	err = json.Unmarshal([]byte(string(body)), &vresp)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	return &vresp, err
}

// 云房估值接口
func valuateHouse(address string, area float64, cityCode string, houseType string) (float64, error) {

	if isMock == "true" {
		totalPrice := 1 * 10000
		return float64(totalPrice), nil
	} else {
		val, err := utilxiaotao.Valuate(address, cityCode, area, houseType)
		if err != nil {
			return 0, err
		}
		// 云房默认返回的数值是以"万元"为单位的，这里更改为以"元"为单位
		totalPrice := val.Price * 10000
		return totalPrice, nil
	}
}
