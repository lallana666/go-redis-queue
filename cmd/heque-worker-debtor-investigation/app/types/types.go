package types

// 查询债务人返回实体
type GetDebtorResponse struct {
	Errors []*ErrorsResponse `json:"errors"`
	Data   *GData            `json:"data"`
}

type ErrorsResponse struct {
	Message *string   `json:"message"`
	Path    []*string `json:"path"`
}

type GData struct {
	EnterpriseBusinessInfo    *EnterpriseBusinessInfo      `json:"enterpriseBusinessInfo"`
	EntZhixingInfo            []*EntZhixingInfo            `json:"entZhixingInfo"`
	EntDiscreditInfo          []*EntDiscreditInfo          `json:"entDiscreditInfo"`
	EntConsumptionRestriction []*EntConsumptionRestriction `json:"entConsumptionRestriction"`
	Holder                    []*Holder                    `json:"Holder"`
	PersonShixin              GetRiskPersonsResponse       `json:"shixin"`
	PersonZx                  GetRiskPersonsResponse       `json:"zx"`
}

// 个人风险信息
type GetRiskPersonsResponse struct {
	//  个人风险信息列表
	AllList []*PersonRiskBody `json:"allList"`
	//  个人失信信息列表
	ShixinList []*PersonRiskBody `json:"shixinList"`
	//  个人执行信息列表
	ZxggList []*PersonRiskBody `json:"zxggList"`
	//  返回code值
	Code string `json:"code"`
	//  返回code值
	Msg *string `json:"msg"`
	//  每页数量
	Range *int `json:"range"`
	//   当前页码数
	PageNo *int `json:"pageNo"`
	//  所有信息数量
	TotalCount *int `json:"totalCount"`
	//  所有信息总页码数
	TotalPageNum *int `json:"totalPageNum"`
	//  检索时间（秒）
	SearchSeconds *float64 `json:"searchSeconds"`
}

type PersonRiskBody struct {
	//  内容概要
	Body *string `json:"body"`
	//  数据类型
	DataType *string `json:"dataType"`
	//  唯一的标识
	EntryID *string `json:"entryId"`
	//  时间
	SortTime *int `json:"sortTime"`
	//  标题
	Title *string `json:"title"`
	//  匹配度
	MatchRatio *float64 `json:"matchRatio"`
	//  执行公告
	Zxgg *ZxggRiskBody `json:"zxgg"`
	//  失信公告
	Shixin *ShixinRiskBody `json:"shixin"`
}

// 执行公告
type ZxggRiskBody struct {
	//  内容概要
	Body *string `json:"body"`
	//  数据类型
	DataType *string `json:"dataType"`
	//  唯一的标识
	ZxggID *string `json:"zxggId"`
	//  时间
	SortTime *int `json:"sortTime"`
	//  标题
	Title *string `json:"title"`
	//  当事人
	Partys []*ZxggPartysRiskBody `json:"partys"`
	//  案号
	CaseNo *string `json:"caseNo"`
	//  法院名称
	Court *string `json:"court"`
	//  地址
	Address *string `json:"address"`
	//  终本日期
	CloseDate *string `json:"closeDate"`
	//  申请人
	Proposer *string `json:"proposer"`
	//  依据文书
	YjCode *string `json:"yjCode"`
	//  依据单位
	Yjdw *string `json:"yjdw"`
}

// 失信公告
type ShixinRiskBody struct {
	//  内容概要
	Body *string `json:"body"`
	//  数据类型
	DataType *string `json:"dataType"`
	//  唯一的标识
	ShixinID *string `json:"shixinId"`
	//  时间
	SortTime *int `json:"sortTime"`
	//  标题
	Title *string `json:"title"`
	//  当事人
	Partys []*ShixinPartysRiskBody `json:"partys"`
	//  案号
	CaseNo *string `json:"caseNo"`
	//  法院名称
	Court *string `json:"court"`
	//  义务
	Yiwu *string `json:"yiwu"`
	//  发布时间
	PostTime *int `json:"postTime"`
	//  依据文号
	YjCode *string `json:"yjCode"`
	//  依据单位
	Yjdw *string `json:"yjdw"`
}

type ZxggPartysRiskBody struct {
	//  案件状态
	CaseStateT *string `json:"caseStateT"`
	//  执行金额
	ExecMoney *int `json:"execMoney"`
	//  身份证号码
	IdcardNo *string `json:"idcardNo"`
	//  主体类型
	PartyType *string `json:"partyType"`
	//  当事人名称
	Pname *string `json:"pname"`
}

type ShixinPartysRiskBody struct {
	//  年龄
	Age *int `json:"age"`
	//  执行金额
	Money *float64 `json:"money"`
	//  身份证号码
	IdcardNo *string `json:"idcardNo"`
	//  具体情形
	Jtqx *string `json:"jtqx"`
	//  履行情况
	LxqkT *string `json:"lxqkT"`
	//  主体类型
	PartyType *string `json:"partyType"`
	//  当事人名称
	Pname *string `json:"pname"`
	//  省份
	Province *string `json:"province"`
}

// EnterpriseBusinessInfo represents 公司工商信息
type EnterpriseBusinessInfo struct {
	ID                    string  `json:"id"`
	StaffNumRange         *string `json:"staffNumRange"`
	FromTime              *int    `json:"fromTime"`
	Type                  *int    `json:"type"`
	BondName              *string `json:"bondName"`
	IsMicroEnt            *int    `json:"isMicroEnt"`
	UsedBondName          *string `json:"usedBondName"`
	RegNumber             *string `json:"regNumber"`
	PercentileScore       *int    `json:"percentileScore"`
	RegCapital            *string `json:"regCapital"`
	Name                  *string `json:"name"`
	RegInstitute          *string `json:"regInstitute"`
	RegLocation           *string `json:"regLocation"`
	Industry              *string `json:"industry"`
	ApprovedTime          *int    `json:"approvedTime"`
	SocialStaffNum        *int    `json:"socialStaffNum"`
	Tags                  *string `json:"tags"`
	TaxNumber             *string `json:"taxNumber"`
	BusinessScope         *string `json:"businessScope"`
	Property3             *string `json:"property3"`
	Alias                 *string `json:"alias"`
	OrgNumber             *string `json:"orgNumber"`
	RegStatus             *string `json:"regStatus"`
	EstiblishTime         *int    `json:"estiblishTime"`
	BondType              *string `json:"bondType"`
	LegalPersonName       *string `json:"legalPersonName"`
	ToTime                *int    `json:"toTime"`
	ActualCapital         *string `json:"actualCapital"`
	CompanyOrgType        *string `json:"companyOrgType"`
	Base                  *string `json:"base"`
	CreditCode            *string `json:"creditCode"`
	HistoryNames          *string `json:"historyNames"`
	BondNum               *string `json:"bondNum"`
	RegCapitalCurrency    *string `json:"regCapitalCurrency"`
	ActualCapitalCurrency *string `json:"actualCapitalCurrency"`
	RevokeDate            *int    `json:"revokeDate"`
	RevokeReason          *string `json:"revokeReason"`
	CancelDate            *int    `json:"cancelDate"`
	CancelReason          *string `json:"cancelReason"`
}

// EntZhixingInfo represents 公司失信信息
type EntZhixingInfo struct {
	CaseCode       *string `json:"caseCode"`
	ExecCourtName  *string `json:"execCourtName"`
	Pname          *string `json:"pname"`
	PartyCardNum   *string `json:"partyCardNum"`
	CaseCreateTime *int    `json:"caseCreateTime"`
	ExecMoney      *string `json:"execMoney"`
}

// EntDiscreditInfo represents 公司失信信息
type EntDiscreditInfo struct {
	BusinessEntity  *string `json:"businessentity"`
	AreaName        *string `json:"areaname"`
	CourtName       *string `json:"courtname"`
	UnPerformPart   *string `json:"unperformPart"`
	Type            *string `json:"type"`
	PerformedPart   *string `json:"performedPart"`
	Iname           *string `json:"iname"`
	DisruptTypeName *string `json:"disrupttypename"`
	CaseCode        *string `json:"casecode"`
	CardNum         *string `json:"cardnum"`
	Performance     *string `json:"performance"`
	RegDate         *int64  `json:"regdate"`
	PublishDate     *int64  `json:"publishdate"`
	GistUnit        *string `json:"gistunit"`
	Duty            *string `json:"duty"`
	GistId          *string `json:"gistid"`
}

// EntConsumptionRestriction represents 公司限制消费信息
type EntConsumptionRestriction struct {
	CaseCode       *string `json:"caseCode"`
	QyinfoAlias    *string `json:"qyinfoAlias"`
	Qyinfo         *string `json:"qyinfo"`
	CaseCreateTime *int    `json:"caseCreateTime"`
	Alias          *string `json:"alias"`
	Xname          *string `json:"xname"`
	FilePath       *string `json:"filePath"`
}

// Holder represents 公司限制消费信息
type Holder struct {
	Name        *string        `json:"name"`
	Alias       *string        `json:"alias"`
	Type        *int           `json:"type"`
	CapitalActl []*CapitalActl `json:"capitalActl"`
	Capital     []*Capital     `json:"capital"`
}

type CapitalActl struct {
	Amomon  *string `json:"amomon"`
	Percent *string `json:"percent"`
	Paymet  *string `json:"paymet"`
}

type Capital struct {
	Amomon  *string `json:"amomon"`
	Percent *string `json:"percent"`
	Paymet  *string `json:"paymet"`
}

// 更新债务人返回实体
type DebtorResponse struct {
	Errors []*ErrorsResponse `json:"errors"`
	Data   *MData            `json:"data"`
}

type MData struct {
	UpdateDebtorInvestigation UpdateDebtorInvestigation `json:"updateDebtorInvestigation"`
}

type UpdateDebtorInvestigation struct {
	DebtorID string  `json:"debtorId"`
	Name     *string `json:"name"`
}
