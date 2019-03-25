package rule

import (
	"io/ioutil"
	"time"

	"github.com/prometheus/prometheus/rules"
	"github.com/prometheus/tsdb"
	"gopkg.in/yaml.v2"
)

type RuleGroups struct {
	PartialResponseDisabled bool `yaml:"partial_response_disabled"`
}

// Update updates files into two different rule managers strict and non strict based on special field in RuleGroup file.
func Update(evalInterval time.Duration, files []string, nonStrict *rules.Manager, strict *rules.Manager) error {
	var (
		errs                       tsdb.MultiError
		strictFile, nonStrictFiles []string
	)
	for _, fn := range files {
		b, err := ioutil.ReadFile(fn)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		var groups RuleGroups
		if err := yaml.UnmarshalStrict(b, &groups); err != nil {
			errs = append(errs, err)
			continue
		}

		if groups.PartialResponseDisabled {
			strictFile = append(strictFile, fn)
		} else {
			nonStrictFiles = append(nonStrictFiles, fn)
		}
	}

	if err := nonStrict.Update(evalInterval, nonStrictFiles); err != nil {
		errs = append(errs, err)
	}

	if err := strict.Update(evalInterval, strictFile); err != nil {
		errs = append(errs, err)
	}

	return errs.Err()
}
