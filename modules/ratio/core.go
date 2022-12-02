package ratio

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/patrickmn/go-cache"
	"gorm.io/gorm"
)

type RatioTable struct {
	RatioName  string `gorm:"type:varchar(50);index;primaryKey"`
	RatioRules string `gorm:"type:varchar(10)"`
}

func (RatioTable) TableName() string {
	return "ratio"
}

type RatioConfig struct {
	Name         string
	DB           *gorm.DB
	SyncInterval time.Duration
}

type RatioCore struct {
	RatioConfig *RatioConfig
	RatioTable  *RatioTable
	Cache       *cache.Cache
	sync.Mutex
}

func NewRatio(config RatioConfig) (*RatioCore, error) {
	var (
		db    = config.DB
		table = &RatioTable{}
		core  = RatioCore{
			RatioConfig: &config,
			RatioTable:  table,
		}
	)

	if config.Name == "" {
		return nil, fmt.Errorf("invalid argument, name cannot be null")
	}

	db.AutoMigrate(table)

	err := db.Where("ratio_name = ?", config.Name).Last(table).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			table.RatioName = config.Name
			table.RatioRules = "1:1"
			db.Create(table)
		} else {
			return nil, err
		}
	}

	if core.RatioTable.RatioRules == "" {
		return nil, fmt.Errorf("invalid argument, rules cannot be null")
	}

	core.Cache = cache.New(0, 0)

	if err := core.SetRules(core.RatioTable.RatioRules); err != nil {
		return nil, err
	}

	go func() {
		for {
			err := db.Where("ratio_name = ?", config.Name).Last(table).Error
			if err == nil {
				core.SetRules(table.RatioRules)
			}
			time.Sleep(config.SyncInterval)
		}
	}()

	return &core, nil
}

func (c *RatioCore) SetCounter(value int) {
	c.Lock()
	c.Cache.Set("counter", value, 0)
	c.Unlock()
}

func (c *RatioCore) GetCounter() int {
	v, found := c.Cache.Get("counter")
	if !found {
		return -1
	}
	return v.(int)
}

func (c *RatioCore) ClearCounter() {
	c.Cache.Set("counter", 0, 0)
}

func (c *RatioCore) SetRules(rules string) error {
	x := strings.Split(rules, ":")
	if len(x) < 2 {
		return fmt.Errorf("invalid rules format, details : %s", rules)
	}

	ruleA, ruleB := x[0], x[1]

	ruleAInt, errRuleA := strconv.Atoi(ruleA)
	ruleBInt, errRuleB := strconv.Atoi(ruleB)
	if errRuleA != nil || errRuleB != nil {
		return fmt.Errorf("invalid rules format, errorRuleA[%s], errorRuleB[%s]", errRuleA, errRuleB)
	}

	c.Cache.Set("rule_a", ruleAInt, 0)
	c.Cache.Set("rule_b", ruleBInt, 0)

	return nil
}

func (c *RatioCore) GetRules() (int, int) {
	a, _ := c.Cache.Get("rule_a")
	b, _ := c.Cache.Get("rule_b")
	c.RatioTable.RatioRules = fmt.Sprintf("%v:%v", a, b)
	return a.(int), b.(int)
}

func (c *RatioCore) IsLimited() bool {
	ra, rb := c.GetRules()

	ctr := c.GetCounter()
	if ctr < 0 {
		c.SetCounter(0)
		ctr = c.GetCounter()
	}

	if ctr >= ra {
		if ctr == rb {
			c.SetCounter(1)
			return false
		}

		c.SetCounter(ctr + 1)
		return true
	}

	c.SetCounter(ctr + 1)
	return false
}
