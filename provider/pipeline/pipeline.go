package pipeline

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"regexp"
	"strings"

	xj "github.com/basgys/goxml2json"
	"github.com/evcc-io/evcc/provider/javascript"
	"github.com/evcc-io/evcc/util/jq"
	"github.com/itchyny/gojq"
	"github.com/robertkrimen/otto"
	"github.com/volkszaehler/mbmd/meters/rs485"
)

type Pipeline struct {
	re     *regexp.Regexp
	jq     *gojq.Query
	unpack string
	decode string
	vm     *otto.Otto
	script string
}

type Settings struct {
	Regex  string
	Jq     string
	Unpack string
	Decode string
	VM     string
	Script string
}

func New(cc Settings) (*Pipeline, error) {
	p := new(Pipeline)

	var err error
	if err == nil && cc.Regex != "" {
		_, err = p.WithRegex(cc.Regex)
	}

	if err == nil && cc.Jq != "" {
		_, err = p.WithJq(cc.Jq)
	}

	if err == nil && cc.Unpack != "" {
		_, err = p.WithUnpack(cc.Unpack)
	}

	if err == nil && cc.Decode != "" {
		_, err = p.WithDecode(cc.Decode)
	}

	if err == nil && cc.Script != "" {
		_, err = p.WithScript(cc.VM, cc.Script)
	}

	return p, err
}

// WithRegex adds a regex query applied to the mqtt listener payload
func (p *Pipeline) WithRegex(regex string) (*Pipeline, error) {
	re, err := regexp.Compile(regex)
	if err != nil {
		return nil, fmt.Errorf("invalid regex '%s': %w", re, err)
	}

	p.re = re

	return p, nil
}

// WithJq adds a jq query applied to the mqtt listener payload
func (p *Pipeline) WithJq(jq string) (*Pipeline, error) {
	op, err := gojq.Parse(jq)
	if err != nil {
		return nil, fmt.Errorf("invalid jq query '%s': %w", jq, err)
	}

	p.jq = op

	return p, nil
}

// WithUnpack adds data unpacking
func (p *Pipeline) WithUnpack(unpack string) (*Pipeline, error) {
	p.unpack = strings.ToLower(unpack)

	return p, nil
}

// WithDecode adds data decoding
func (p *Pipeline) WithDecode(decode string) (*Pipeline, error) {
	p.decode = strings.ToLower(decode)

	return p, nil
}

// WithScript adds a javascript script to process the response
func (p *Pipeline) WithScript(vm, script string) (*Pipeline, error) {
	regvm := javascript.RegisteredVM(strings.ToLower(vm))

	p.vm = regvm
	p.script = script

	return p, nil
}

// transform XML into JSON with attribute names getting 'attr' prefix
func (p *Pipeline) transformXML(value []byte) []byte {
	value = bytes.TrimSpace(value)

	// only do a simple check, as some devices e.g. Kostal Piko MP plus don't seem to send proper XML
	if !bytes.HasPrefix(value, []byte("<?xml")) {
		return value
	}

	in := bytes.NewReader(value)

	// Decode XML document
	root := new(xj.Node)
	if err := xj.NewDecoder(in).DecodeWithCustomPrefixes(root, "", "attr"); err != nil {
		return value
	}

	// Then encode it in JSON
	out := new(bytes.Buffer)
	if err := xj.NewEncoder(out).Encode(root); err != nil {
		return value
	}

	return out.Bytes()
}

func (p *Pipeline) unpackValue(value []byte) (string, error) {
	switch p.unpack {
	case "hex":
		b, err := hex.DecodeString(string(value))
		if err != nil {
			return "", err
		}
		return string(b), nil
	}

	return "", fmt.Errorf("invalid unpack: %s", p.unpack)
}

// decode a hex string to a proper value
// TODO reuse similar code from Modbus
func (p *Pipeline) decodeValue(value []byte) (interface{}, error) {
	switch p.decode {
	case "float32", "ieee754":
		return rs485.RTUIeee754ToFloat64(value), nil
	case "float32s", "ieee754s":
		return rs485.RTUIeee754ToFloat64Swapped(value), nil
	case "float64":
		return rs485.RTUUint64ToFloat64(value), nil
	case "uint16":
		return rs485.RTUUint16ToFloat64(value), nil
	case "uint32":
		return rs485.RTUUint32ToFloat64(value), nil
	case "uint32s":
		return rs485.RTUUint32ToFloat64Swapped(value), nil
	case "uint64":
		return rs485.RTUUint64ToFloat64(value), nil
	case "int16":
		return rs485.RTUInt16ToFloat64(value), nil
	case "int32":
		return rs485.RTUInt32ToFloat64(value), nil
	case "int32s":
		return rs485.RTUInt32ToFloat64Swapped(value), nil
	}

	return nil, fmt.Errorf("invalid decoding: %s", p.decode)
}

func (p *Pipeline) Process(in []byte) ([]byte, error) {
	b := p.transformXML(in)

	if p.re != nil {
		m := p.re.FindSubmatch(b)
		if len(m) == 1 {
			b = m[0] // full match
		} else if len(m) > 1 {
			b = m[1] // first submatch
		}
	}

	if p.jq != nil {
		v, err := jq.Query(p.jq, b)
		if err != nil {
			return b, err
		}
		b = []byte(fmt.Sprintf("%v", v))
	}

	if p.unpack != "" {
		v, err := p.unpackValue(b)
		if err != nil {
			return b, err
		}
		b = []byte(fmt.Sprintf("%v", v))
	}

	if p.decode != "" {
		v, err := p.decodeValue(b)
		if err != nil {
			return b, err
		}
		b = []byte(fmt.Sprintf("%v", v))
	}

	if p.vm != nil {
		if err := p.vm.Set("val", string(b)); err != nil {
			return b, err
		}

		v, err := p.vm.Eval(p.script)
		if err != nil {
			return b, err
		}

		s, err := v.ToString()
		b = []byte(s)
		if err != nil {
			return b, err
		}
	}

	return b, nil
}
