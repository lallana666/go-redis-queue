package types

// 查询速销率返回实体
type GetValuationResponse struct {
	Errors []*ErrorsResponse `json:"errors"`
	Data   *GData            `json:"data"`
}

type ErrorsResponse struct {
	Message *string   `json:"message"`
	Path    []*string `json:"path"`
}

type GData struct {
	Viewer Viewer `json:"viewer"`
}

type Viewer struct {
	Org Org `json:"org"`
}

type Org struct {
	Pkg Pkg `json:"pkg"`
}

type Pkg struct {
	Property Property `json:"property"`
}

type Property struct {
	Valuation *string `json:"valuation"`
	Info      Info    `json:"info"`
}

type Info struct {
	Valuation Valuation `json:"valuation"`
}

type Valuation struct {
	RecoveryRate *float64 `json:"recoveryRate"`
}

// 更新估值返回实体
type ValuationResponse struct {
	Errors []*ErrorsResponse `json:"errors"`
	Data   *MData            `json:"data"`
}

type MData struct {
	UpdateHouseValuation UpdateHouseValuation `json:"updateHouseValuation"`
}

type UpdateHouseValuation struct {
	ValuationAmount *float64 `json:"valuationAmount"`
}
