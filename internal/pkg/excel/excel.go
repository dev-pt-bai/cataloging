package excel

import (
	"io"

	"github.com/dev-pt-bai/cataloging/internal/pkg/errors"
	"github.com/xuri/excelize/v2"
)

type Parser struct{}

func NewParser() *Parser { return &Parser{} }

func (p *Parser) Open(r io.Reader) ([][]string, *errors.Error) {
	file, err := excelize.OpenReader(r)
	if err != nil {
		return nil, errors.New(errors.ParsingFileFailure).Wrap(err)
	}
	defer file.Close()

	rows, err := file.GetRows(file.GetSheetName(file.GetActiveSheetIndex()))
	if err != nil {
		return nil, errors.New(errors.ParsingFileFailure).Wrap(err)
	}

	return rows, nil
}
