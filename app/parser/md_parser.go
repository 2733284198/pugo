package parser

import (
	"bufio"
	"bytes"
	"errors"
	"io"
)

const (
	// MdParserPrefix is prefix of MdParser
	MdParserPrefix = "```"
)

type (
	// MdParser parses content as similar markdown content
	MdParser struct {
		blocks []Block
	}
)

// NewMdParser returns new MdParser
func NewMdParser() *MdParser {
	return &MdParser{
		blocks: []Block{new(IniBlock), new(MarkdownBlock)},
	}
}

// Is checks data can be parsed by MdParser
func (mp *MdParser) Is(data []byte) bool {
	data = bytes.TrimLeft(data, "\n")
	return bytes.HasPrefix(data, []byte(MdParserPrefix))
}

// Detect detects block to save bytes
func (mp *MdParser) Detect(mark []byte) Block {
	for _, b := range mp.blocks {
		if b.Is(mark) {
			return b.New()
		}
	}
	return nil
}

// Parse parses bytes to blocks
func (mp *MdParser) Parse(src []byte) ([]Block, error) {
	if len(src) == 0 {
		return nil, nil
	}
	buf := bytes.NewBuffer(src)
	return mp.ParseReader(buf)
}

// ParseReader parsers Reader to blocks
func (mp *MdParser) ParseReader(r io.Reader) ([]Block, error) {
	var (
		currentBlock Block
		blocks       []Block
		reader       = bufio.NewReader(r)
		isLined      bool
	)
	for {
		lineData, _, err := reader.ReadLine()
		if currentBlock == nil {
			if len(lineData) == 0 {
				continue
			}
			if len(blocks) == 0 {
				// the first block need be MetaBlock
				if currentBlock = mp.Detect(bytes.TrimLeft(lineData, MdParserPrefix)); currentBlock == nil {
					return nil, errors.New("block-parse-first-error")
				}
				continue
			}
		}

		// when parsing first block, check end mark to close the block.
		if len(blocks) == 0 && bytes.Equal(lineData, []byte(MdParserPrefix)) {
			blocks = append(blocks, currentBlock)
			currentBlock = new(MarkdownBlock).New()
			continue
		}

		// write block
		if len(lineData) > 0 || isLined {
			if err := currentBlock.Write(append(lineData, []byte("\n")...)); err != nil {
				return nil, err
			}
			if len(blocks) > 0 {
				isLined = true
			}
		}

		if err != nil {
			if err != io.EOF {
				return nil, err
			}
			break
		}
	}

	// this block must be md block,
	// if empty, do not use it, so same behavior as without -----markdown block
	if currentBlock != nil && len(currentBlock.Bytes()) > 0 {
		blocks = append(blocks, currentBlock)
	}
	return blocks, nil
}
