package app

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/golang/glog"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"denggotech.cn/heque/heque/client"
	"denggotech.cn/heque/heque/cmd/heque-worker-debtor-investigation/app/types"
	utilflag "denggotech.cn/heque/heque/util/flag"
)

// 人法尽调消费实体类
type jobArgs struct {
	//  pkg id
	PkgID string `json:"pkgID"`
	//  The debtor's name.
	Name string `json:"name"`
	//  估值token
	Token string `json:"token"`
	//  批次id
	Batch             string    `json:"batch"`
	Kind              string    `json:"kind"`
	DebtorID          string    `json:"debtorID"`
	IDNumber          string    `json:"idNumber"`
	Birthplace        string    `json:"birthplace"`
	Birthday          time.Time `json:"birthday"`
	BusinessTimeLimit time.Time `json:"businessTimeLimit"`
	BusinessTimeTo    time.Time `json:"businessTimeTo"`
	RegisterPlace     string    `json:"registerPlace"`
	PayCapital        float64   `json:"payCapital"`
	Currency          string    `json:"currency"`
	StaffNum          string    `json:"staffNum"`
	InsuredNum        string    `json:"insuredNum"`
	CompanyIndustry   string    `json:"companyIndustry"`
	CompanyArea       string    `json:"companyArea"`
	RegisterCapital   float64   `json:"registerCapital"`
	ApprovedAt        time.Time `json:"approvedAt"`
	BusinessScope     string    `json:"businessScope"`
}

func NewWorkerCommand() *cobra.Command {
	s := NewWorkerOptions()

	cmd := &cobra.Command{
		Use:  "heque-worker-debtor-investigation",
		Long: `This worker investigates debtor.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			utilflag.PrintFlags(cmd.Flags())

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
	// Initialize the credit-gateway ismock
	err := initIsMock(w.IsMock)
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
			glog.Errorf("Observed a error: %s", err)
			continue
		}

		err = consumeOneJob(job, w.DebtdbAddress, w.CreditGatewayAddress, cli)
		if err == nil {
			cli.MarkAsDone(job)
		} else {
			glog.Errorf("consume job failed: %s", err)
			cli.MarkAsFailed(job)
		}
	}

	return nil
}

func consumeOneJob(j *client.Job, debtdbAdress string, creditGatewayAddress string, cli *client.Client) error {
	glog.Infof("consuming job: %s", j.ID)

	var jobArgs jobArgs

	err := json.Unmarshal([]byte(j.Spec.Payload), &jobArgs)
	if err != nil {
		return err
	}

	if jobArgs.Kind == "121" { // 个人：121
		// 向信用网关查询人法信息
		payloadStrQueryVal := fmt.Sprintf("{\"operationName\":null,\"variables\":{},\"query\":\"{shixin:riskPersons(name:\\\"%s\\\", idcardNo: \\\"%s\\\", domain: \\\"sifa\\\", dataType: \\\"shixin\\\") {     code     msg     shixinList{       body       dataType       entryId       sortTime       title       matchRatio       shixin{         shixinId         body         caseNo         court         postTime         sortTime         yiwu         yjCode         yjdw         dataType       }     }   }   zx:riskPersons(name: \\\"%s\\\", idcardNo: \\\"%s\\\", domain: \\\"sifa\\\", dataType: \\\"zxgg\\\") {     code     msg     zxggList{       body       dataType       entryId       sortTime       title       matchRatio       zxgg{         zxggId         address         body         caseNo         closeDate         court         proposer         sortTime         title         yjCode         yjdw       }     }   }  }\"}",
			jobArgs.Name, jobArgs.IDNumber, jobArgs.Name, jobArgs.IDNumber)

		getDebtorResponse, err := httpGraphqlInvestigationQuery(&jobArgs, payloadStrQueryVal, creditGatewayAddress)
		if err != nil {
			return err
		}
		if getDebtorResponse.Errors != nil {
			return errors.New("查询人法信息出错")
		}

		// graphql 写回debtdb
		payloadStrUpdateVal := "{\"operationName\":null,\"variables\":{\"input\":{\"debtorId\":\"%s\",\"name\":\"%s\","
		args := []interface{}{jobArgs.DebtorID, jobArgs.Name}

		if getDebtorResponse.Data.PersonShixin.Code == "s" {
			payloadStrUpdateVal += "\"discredits\":["
			for _, shixinList := range getDebtorResponse.Data.PersonShixin.ShixinList {
				payloadStrUpdateVal += "\"entryId\":\"%s\","
				args = append(args, *shixinList.EntryID)
				var sortTime string
				var postTime string
				if shixinList.SortTime != nil {
					payloadStrUpdateVal += "\"sortTime\":\"%s\","
					sortTime = time.Unix(0, int64(*shixinList.SortTime)*1e6).Format("2006-01-02T15:04:05Z07:00")
					args = append(args, sortTime)
				}
				if shixinList.Shixin.PostTime != nil {
					payloadStrUpdateVal += "\"postTime\":\"%s\","
					postTime = time.Unix(0, int64(*shixinList.Shixin.PostTime)*1e6).Format("2006-01-02T15:04:05Z07:00")
					args = append(args, postTime)
				}

				payloadStrUpdateVal += "{\"title\":\"%s\",\"caseNo\":\"%s\",\"court\":\"%s\",\"yjCode\":\"%s\",\"yjdw\":\"%s\",\"yiwu\":\"%s\",\"body\":\"%s\"},"
				args = append(args, *shixinList.Title, *shixinList.Shixin.CaseNo, *shixinList.Shixin.Court, *shixinList.Shixin.YjCode, *shixinList.Shixin.Yjdw, *shixinList.Shixin.Yiwu, *shixinList.Shixin.Body)
			}
			payloadStrUpdateVal = payloadStrUpdateVal[0:len(payloadStrUpdateVal)-1] + "]"

		}
		if getDebtorResponse.Data.PersonZx.Code == "s" {
			payloadStrUpdateVal += ",\"zhixing\":["
			for _, zxggList := range getDebtorResponse.Data.PersonZx.ZxggList {
				payloadStrUpdateVal += "{\"entryId\":\"%s\","
				args = append(args, *zxggList.EntryID)

				var sortTime string
				if zxggList.SortTime != nil {
					payloadStrUpdateVal += "\"sortTime\":\"%s\","
					sortTime = time.Unix(0, int64(*zxggList.SortTime)*1e6).Format("2006-01-02T15:04:05Z07:00")
					args = append(args, sortTime)
				}

				payloadStrUpdateVal += "{\"title\":\"%s\",\"caseNo\":\"%s\",\"court\":\"%s\",\"yjCode\":\"%s\",\"yjdw\":\"%s\",\"proposer\":\"%s\",\"address\":\"%s\",\"body\":\"%s\"},"
				args = append(args, *zxggList.Title, *zxggList.Zxgg.CaseNo, *zxggList.Zxgg.Court, *zxggList.Zxgg.YjCode, *zxggList.Zxgg.Yjdw, *zxggList.Zxgg.Proposer, *zxggList.Zxgg.Address, *zxggList.Zxgg.Body)
			}
			payloadStrUpdateVal = payloadStrUpdateVal[0:len(payloadStrUpdateVal)-1] + "]"
		}
		payloadStrUpdateVal = fmt.Sprintf(payloadStrUpdateVal, args...)
		payloadStrUpdateVal = payloadStrUpdateVal + "}},\"query\": \"mutation ($input: UpdateDebtorInput!) {updateDebtor(input: $input) { id }}\"}"

		updateDebtorResponse, err := httpGraphqlInvestigationMutation(&jobArgs, payloadStrUpdateVal, debtdbAdress)
		if err != nil {
			return err
		}
		if updateDebtorResponse.Errors != nil {
			return errors.New("更新债务人出错")
		}
	} else if jobArgs.Kind == "122" { // 公司：122
		// 向信用网关查询人法信息
		payloadStrQueryVal := fmt.Sprintf("{\"operationName\":null,\"variables\":{},\"query\":\"{\\n  enterpriseBusinessInfo(keyword: \\\"%s\\\") {\\n    id\\n    name\\n    staffNumRange\\n    fromTime\\n    type\\n    bondName\\n    isMicroEnt\\n    usedBondName\\n    regNumber\\n    percentileScore\\n    regCapital\\n    regInstitute\\n    regLocation\\n    industry\\n    approvedTime\\n    socialStaffNum\\n    tags\\n    taxNumber\\n    businessScope\\n    property3\\n    alias\\n    orgNumber\\n    regStatus\\n    estiblishTime\\n    bondType\\n    legalPersonName\\n    toTime\\n    actualCapital\\n    companyOrgType\\n    base\\n    creditCode\\n    historyNames\\n    bondNum\\n    regCapitalCurrency\\n    actualCapitalCurrency\\n    revokeDate\\n    revokeReason\\n    cancelDate\\n    cancelReason\\n  }\\n  "+
			"entDiscreditInfo(name: \\\"%s\\\") {\\n    businessEntity\\n    areaName\\n    courtName\\n    unPerformPart\\n    type\\n    performedPart\\n    iname\\n    disruptTypeName\\n    caseCode\\n    cardNum\\n    performance\\n    regDate\\n    publishDate\\n    gistUnit\\n    duty\\n    gistId\\n  }\\n  "+
			"entZhixingInfo(name: \\\"%s\\\") {\\n    caseCode\\n    execCourtName\\n    pname\\n    partyCardNum\\n    caseCreateTime\\n    execMoney\\n  }\\n  "+
			"entConsumptionRestriction(name: \\\"%s\\\") {\\n    caseCode\\n    qyinfoAlias\\n    qyinfo\\n    caseCreateTime\\n    alias\\n    xname\\n    filePath\\n  }\\n  "+
			"holder(name: \\\"%s\\\") {\\n    name\\n    alias\\n    capitalActl {\\n      amomon\\n      percent\\n    }\\n    capital {\\n      amomon\\n      percent\\n    }\\n    type\\n  }\\n}\\n\"}",
			jobArgs.Name, jobArgs.Name, jobArgs.Name, jobArgs.Name, jobArgs.Name)

		getDebtorResponse, err := httpGraphqlInvestigationQuery(&jobArgs, payloadStrQueryVal, creditGatewayAddress)
		if err != nil {
			return err
		}
		if getDebtorResponse.Errors != nil {
			return errors.New("查询人法信息出错")
		}

		// graphql 写回debtdb
		payloadStrUpdateVal := "{\"operationName\":null,\"variables\":{\"input\":{\"debtorId\":\"%s\",\"name\":\"%s\""
		args := []interface{}{jobArgs.DebtorID, jobArgs.Name}

		// 拼接enterpriseBusinessInfo
		entBusinessInfo := getDebtorResponse.Data.EnterpriseBusinessInfo
		if entBusinessInfo != nil {
			payloadStrUpdateVal += ",\"idNumber\":\"%s\",\"legalPerson\":\"%s\",\"staffNum\":\"%s\",\"insuredNum\":\"%d\",\"companyIndustry\":\"%s\",\"companyArea\":\"%s\",\"registerCapital\":\"%f\",\"payCapital\":\"%f\",\"currency\":\"%s\","

			var regCapital float64
			if entBusinessInfo.RegCapital != nil {
				s := splitNumAndChar(*entBusinessInfo.RegCapital)[0]
				float, err := strconv.ParseFloat(s, 64)
				if err != nil {
					return err
				}
				regCapital = float
			}
			var payCapital float64
			if entBusinessInfo.ActualCapital != nil {
				s := splitNumAndChar(*entBusinessInfo.RegCapital)[0]
				float, err := strconv.ParseFloat(s, 64)
				if err != nil {
					return err
				}
				payCapital = float
			}
			var currency string
			if entBusinessInfo.RegCapital != nil {
				s := splitNumAndChar(*entBusinessInfo.RegCapital)[1]
				currency = s
			} else {
				currency = "万人民币"
			}
			args = append(args, *entBusinessInfo.CreditCode, *entBusinessInfo.LegalPersonName, *entBusinessInfo.StaffNumRange, *entBusinessInfo.SocialStaffNum, *entBusinessInfo.Industry, *entBusinessInfo.RegLocation, regCapital, payCapital, currency)

			var birthday string
			var approvedAt string
			var businessTimeLimit string
			var businessTimeTo string
			if entBusinessInfo.EstiblishTime != nil {
				payloadStrUpdateVal += "\"birthday\":\"%s\","
				birthday = time.Unix(0, int64(*entBusinessInfo.EstiblishTime)*1e6).Format("2006-01-02T15:04:05Z07:00")
				args = append(args, birthday)
			}
			if entBusinessInfo.ApprovedTime != nil {
				payloadStrUpdateVal += "\"approvedAt\":\"%s\","
				approvedAt = time.Unix(0, int64(*entBusinessInfo.ApprovedTime)*1e6).Format("2006-01-02T15:04:05Z07:00")
				args = append(args, approvedAt)
			}
			if entBusinessInfo.ApprovedTime != nil {
				payloadStrUpdateVal += "\"businessTimeLimit\":\"%s\","
				businessTimeLimit = time.Unix(0, int64(*entBusinessInfo.ApprovedTime)*1e6).Format("2006-01-02T15:04:05Z07:00")
				args = append(args, businessTimeLimit)
			}
			if entBusinessInfo.ToTime != nil {
				payloadStrUpdateVal += "\"businessTimeTo\":\"%s\","
				businessTimeTo = time.Unix(0, int64(*entBusinessInfo.ToTime)*1e6).Format("2006-01-02T15:04:05Z07:00")
				args = append(args, businessTimeTo)
			}

			payloadStrUpdateVal += "\"birthplace\":\"%s\",\"registerPlace\":\"%s\",\"businessScope\":\"%s\""
			args = append(args, *entBusinessInfo.RegLocation, *entBusinessInfo.RegInstitute, *entBusinessInfo.BusinessScope)
		}

		// 拼接shareholderEntities
		if getDebtorResponse.Data.Holder != nil {
			payloadStrUpdateVal += ",\"shareholderEntities\":["
			for _, holder := range getDebtorResponse.Data.Holder {
				var capitalContributions = 0.0
				var realCapitalContributions = 0.0
				var rate = 0.0
				var realRate = 0.0
				for _, capital := range holder.Capital {
					if capital.Percent != nil {
						s := *capital.Percent
						str := s[0 : len(s)-1]
						float, err := strconv.ParseFloat(str, 64)
						if err != nil {
							return err
						}
						rate += float
					}
					if capital.Amomon != nil {
						s := splitNumAndChar(*capital.Amomon)[0]
						float, err := strconv.ParseFloat(s, 64)
						if err != nil {
							return err
						}
						capitalContributions += float
					}
				}
				for _, capital := range holder.CapitalActl {
					if capital.Percent != nil {
						s := *capital.Percent
						str := s[0 : len(s)-1]
						float, err := strconv.ParseFloat(str, 64)
						if err != nil {
							return err
						}
						realRate += float
					}
					if capital.Amomon != nil {
						s := splitNumAndChar(*capital.Amomon)[0]
						float, err := strconv.ParseFloat(s, 64)
						if err != nil {
							return err
						}
						realCapitalContributions += float
					}
				}
				payloadStrUpdateVal += "{\"name\":\"%s\",\"rate\":\"%f\",\"capitalContributions\":\"%f\",\"realCapitalContributions\":\"%f\"},"
				if realRate == 0.0 {
					realRate = rate
				}
				if realCapitalContributions == 0.0 {
					realCapitalContributions = capitalContributions
				}
				args = append(args, *holder.Name, realRate, capitalContributions, realCapitalContributions)
			}
			payloadStrUpdateVal = payloadStrUpdateVal[0:len(payloadStrUpdateVal)-1] + "]"
		}

		// 拼接EntDiscreditInfo
		if getDebtorResponse.Data.EntDiscreditInfo != nil {
			payloadStrUpdateVal += ",\"discredits\":["
			for _, discreditInfo := range getDebtorResponse.Data.EntDiscreditInfo {
				payloadStrUpdateVal += "{\"businessEntity\":\"%s\",     \"areaName\":\"%s\",     \"courtName\":\"%s\",     \"unPerformPart\":\"%s\","

				var businessEntity string
				var areaName string
				var courtName string
				var unPerformPart string
				if discreditInfo.BusinessEntity != nil {
					businessEntity = *discreditInfo.BusinessEntity
				}
				if discreditInfo.AreaName != nil {
					areaName = *discreditInfo.AreaName
				}
				if discreditInfo.CourtName != nil {
					courtName = *discreditInfo.CourtName
				}
				if discreditInfo.UnPerformPart != nil {
					unPerformPart = *discreditInfo.UnPerformPart
				}
				args = append(args, businessEntity, areaName, courtName, unPerformPart)
				payloadStrUpdateVal += "\"type\":\"%s\",\"performedPart\":\"%s\",\"iname\":\"%s\",\"disruptTypeName\":\"%s\",\"caseCode\":\"%s\",\"cardNum\":\"%s\",\"performance\":\"%s\",\"regDate\":\"%s\",\"publishDate\":\"%s\",\"gistUnit\":\"%s\",\"duty\":\"%s\",\"gistId\":\"%s\"},"

				regDate := time.Unix(0, *discreditInfo.RegDate*1e6).Format("2006-01-02T15:04:05Z07:00")
				publishDate := time.Unix(0, *discreditInfo.PublishDate*1e6).Format("2006-01-02T15:04:05Z07:00")
				// duty字段出现了非法字符，如\r,这里进行转换
				*discreditInfo.Duty = strings.Replace(*discreditInfo.Duty, "\r", "", -1)
				*discreditInfo.Duty = strings.Replace(*discreditInfo.Duty, "\n", "", -1)

				args = append(args, *discreditInfo.Type, *discreditInfo.PerformedPart, *discreditInfo.Iname, *discreditInfo.DisruptTypeName, *discreditInfo.CaseCode, *discreditInfo.CardNum, *discreditInfo.Performance, regDate, publishDate, *discreditInfo.GistUnit, *discreditInfo.Duty, *discreditInfo.GistId)
			}
			payloadStrUpdateVal = payloadStrUpdateVal[0:len(payloadStrUpdateVal)-1] + "]"
		}

		// 拼接EntZhixingInfo
		if getDebtorResponse.Data.EntZhixingInfo != nil {
			payloadStrUpdateVal += ",\"zhixing\":["
			for _, entZhixingInfo := range getDebtorResponse.Data.EntZhixingInfo {
				payloadStrUpdateVal += "{\"caseCode\":\"%s\",\"execCourtName\":\"%s\",\"pname\":\"%s\",\"partyCardNum\":\"%s\",\"execMoney\":\"%s\","
				args = append(args, *entZhixingInfo.CaseCode, *entZhixingInfo.ExecCourtName, *entZhixingInfo.Pname, *entZhixingInfo.PartyCardNum, *entZhixingInfo.ExecMoney)
				var caseCreateTime string
				if entZhixingInfo.CaseCreateTime != nil {
					payloadStrUpdateVal += "\"caseCreateTime\":\"%s\""
					caseCreateTime = time.Unix(0, int64(*entZhixingInfo.CaseCreateTime)*1e6).Format("2006-01-02T15:04:05Z07:00")
					args = append(args, caseCreateTime)
				}
				payloadStrUpdateVal += "},"
			}
			payloadStrUpdateVal = payloadStrUpdateVal[0:len(payloadStrUpdateVal)-1] + "]"
		}

		// 拼接consumptionRestriction
		if getDebtorResponse.Data.EntConsumptionRestriction != nil {
			payloadStrUpdateVal += ",\"consumptionRestriction\":["
			for _, consumptionRestriction := range getDebtorResponse.Data.EntConsumptionRestriction {
				payloadStrUpdateVal += "{\"caseCode\":\"%s\",\"qyinfoAlias\":\"%s\",\"qyinfo\":\"%s\",\"alias\":\"%s\",\"xname\":\"%s\",\"filePath\":\"%s\","
				var alias string
				if consumptionRestriction.Alias != nil {
					alias = *consumptionRestriction.Alias
				}
				args = append(args, *consumptionRestriction.CaseCode, *consumptionRestriction.QyinfoAlias, *consumptionRestriction.Qyinfo, alias, *consumptionRestriction.Xname, *consumptionRestriction.FilePath)
				var caseCreateTime string
				if consumptionRestriction.CaseCreateTime != nil {
					payloadStrUpdateVal += "\"caseCreateTime\":\"%s\""
					caseCreateTime = time.Unix(0, int64(*consumptionRestriction.CaseCreateTime)*1e6).Format("2006-01-02T15:04:05Z07:00")
					args = append(args, caseCreateTime)
				}
				payloadStrUpdateVal += "},"
			}
			payloadStrUpdateVal = payloadStrUpdateVal[0 : len(payloadStrUpdateVal)-1]
		}
		payloadStrUpdateVal = fmt.Sprintf(payloadStrUpdateVal, args...)
		payloadStrUpdateVal = payloadStrUpdateVal + "]}},\"query\": \"mutation ($input: UpdateDebtorInput!) {updateDebtor(input: $input) { id }}\"}"

		updateDebtorResponse, err := httpGraphqlInvestigationMutation(&jobArgs, payloadStrUpdateVal, debtdbAdress)
		if err != nil {
			return err
		}
		if updateDebtorResponse.Errors != nil {
			return errors.New("更新债务人出错")
		}
	} else {
		return errors.New("债务人类型错误")
	}

	glog.Infof("debtor investigation is done......jobId:" + j.ID)
	return nil
}

func splitNumAndChar(str string) []string {
	var arr []string
	var num string
	var zwChar string
	for _, s := range str {
		println(s)
		if s < 256 {
			num = num + string(s)
		}

		if s >= 256 {
			zwChar = zwChar + string(s)
		}
	}

	if num != "" {
		arr = append(arr, num)
	}
	if zwChar != "" {
		arr = append(arr, zwChar)
	}
	return arr
}

func httpGraphqlInvestigationMutation(j *jobArgs, payloadStr string, url string) (*types.DebtorResponse, error) {
	payload := strings.NewReader(payloadStr)
	client := &http.Client{}
	req, err := http.NewRequest("POST", url, payload)
	if err != nil {
		glog.Errorf("Observed a error :%s", err)
		return nil, err
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer "+j.Token)

	res, err := client.Do(req)
	if err != nil {
		glog.Errorf("Observed a error :%s", err)
		return nil, err
	}

	if res.StatusCode != http.StatusOK {
		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return nil, err
		}
		glog.Info(string(body))
		return nil, errors.New("更新债务人出错")
	}

	var dr types.DebtorResponse
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		glog.Info(err)
		return nil, err
	}

	err = json.Unmarshal([]byte(string(body)), &dr)
	if err != nil {
		glog.Errorf("Observed a error :%s", err)
		return nil, err
	}

	return &dr, err
}

func httpGraphqlInvestigationQuery(j *jobArgs, payloadStr string, url string) (*types.GetDebtorResponse, error) {
	payload := strings.NewReader(payloadStr)
	client := &http.Client{}
	req, err := http.NewRequest("POST", url, payload)

	if err != nil {
		glog.Errorf("Observed a error :%s", err)
		return nil, err
	}

	req.Header.Add("Content-Type", "application/json")

	res, err := client.Do(req)
	if err != nil {
		glog.Errorf("Observed a error :%s", err)
		return nil, err
	}

	if res.StatusCode != http.StatusOK {
		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return nil, err
		}
		glog.Info(string(body))
		return nil, errors.New("查询债务人报错")
	}

	var vresp types.GetDebtorResponse
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		glog.Errorf("Observed a error :%s", err)
		return nil, err
	}

	err = json.Unmarshal([]byte(string(body)), &vresp)
	if err != nil {
		glog.Errorf("Observed a error :%s", err)
		return nil, err
	}

	return &vresp, err
}
