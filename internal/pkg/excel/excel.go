package excel

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"strconv"

	"github.com/dev-pt-bai/cataloging/internal/pkg/errors"
)

type Parser struct{}

func NewParser() *Parser { return &Parser{} }

func (p *Parser) Open(r io.Reader) ([][]string, *errors.Error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, errors.New(errors.ParsingFileFailure).Wrap(err)
	}

	zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return nil, errors.New(errors.ParsingFileFailure).Wrap(err)
	}

	var sharedStringsFile, sheetFile *zip.File
	for _, f := range zr.File {
		if sharedStringsFile != nil && sheetFile != nil {
			break
		}
		switch f.Name {
		case "xl/sharedStrings.xml":
			sharedStringsFile = f
		case "xl/worksheets/sheet1.xml":
			sheetFile = f
		}
	}

	sharedStrings, err := parseSharedStrings(sharedStringsFile)
	if err != nil {
		return nil, errors.New(errors.ParsingFileFailure).Wrap(err)
	}

	sheet, err := parseSheet(sheetFile, sharedStrings)
	if err != nil {
		return nil, errors.New(errors.ParsingFileFailure).Wrap(err)
	}

	return sheet, nil
}

type sharedStrings struct {
	XMLName xml.Name     `xml:"sst"`
	Items   []sharedItem `xml:"si"`
}

type sharedItem struct {
	T string `xml:"t"`
}

func parseSharedStrings(zf *zip.File) ([]string, error) {
	if zf == nil {
		return nil, nil
	}

	rc, err := zf.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open zip file of shared strings: %w", err)
	}
	defer rc.Close()

	ss := new(sharedStrings)
	if err := xml.NewDecoder(rc).Decode(ss); err != nil {
		return nil, fmt.Errorf("failed to decode xml of shared strings: %w", err)
	}

	values := make([]string, len(ss.Items))
	for i, item := range ss.Items {
		values[i] = item.T
	}

	return values, nil
}

type cell struct {
	R string `xml:"r,attr"`
	T string `xml:"t,attr"`
	V string `xml:"v"`
}

type row struct {
	Cells []cell `xml:"c"`
}

type sheet struct {
	Rows []row `xml:"sheetData>row"`
}

func parseSheet(zf *zip.File, shared []string) ([][]string, error) {
	if zf == nil {
		return nil, nil
	}

	rc, err := zf.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open zip file of sheet: %w", err)
	}
	defer rc.Close()

	s := new(sheet)
	if err := xml.NewDecoder(rc).Decode(s); err != nil {
		return nil, fmt.Errorf("failed to decode xml of sheet: %w", err)
	}

	rows := make([][]string, len(s.Rows))
	for i, row := range s.Rows {
		rows[i] = make([]string, len(row.Cells))
		for j, cell := range row.Cells {
			value := cell.V
			if cell.T == "s" {
				idx, err := strconv.Atoi(cell.V)
				if err == nil && idx < len(shared) {
					value = shared[idx]
				}
			}
			rows[i][j] = value
		}
	}

	return rows, nil
}
