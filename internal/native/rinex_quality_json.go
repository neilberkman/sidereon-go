package native

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
)

type rinexLintFindingDetail struct {
	Kind    string         `json:"kind"`
	SpecRef string         `json:"spec_ref"`
	Details map[string]any `json:"details"`
}

func decodeRINEXLintFindingDetail(payload []byte) (rinexLintFindingDetail, error) {
	decoder := json.NewDecoder(bytes.NewReader(payload))
	decoder.UseNumber()
	var detail rinexLintFindingDetail
	if err := decoder.Decode(&detail); err != nil {
		return rinexLintFindingDetail{}, err
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		if err == nil {
			return rinexLintFindingDetail{}, errors.New("sidereon: trailing RINEX finding detail JSON")
		}
		return rinexLintFindingDetail{}, err
	}
	if detail.Kind == "" || detail.SpecRef == "" || detail.Details == nil {
		return rinexLintFindingDetail{}, errors.New("sidereon: incomplete RINEX finding detail JSON")
	}
	return detail, nil
}
