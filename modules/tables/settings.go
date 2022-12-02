package table

import (
	"net/url"

	"gorm.io/gorm"
)

type Settings struct {
	ID    uint   `gorm:"primarykey"`
	Name  string `gorm:"type:varchar(100)"`
	Value string
}

func (c Settings) GetValue(key string) string {
	data, err := url.ParseQuery(c.Value)
	if err != nil {
		return ""
	}
	return data.Get(key)
}

func (c Settings) GetValueMap() map[string]string {
	data := make(map[string]string)
	values, err := url.ParseQuery(c.Value)
	if err != nil {
		return data
	}
	for k, v := range values {
		data[k] = v[0]
	}
	return data
}

func (c Settings) SetData(db *gorm.DB, key, value string) error {
	values, err := url.ParseQuery(c.Value)
	if err != nil {
		return err
	}
	values.Set(key, value)
	err = db.Model(&Settings{}).Where("id = ?", c.ID).Update("value", values.Encode()).Error
	if err != nil {
		return err
	}
	return nil
}

func (c Settings) SetDataMap(db *gorm.DB, data map[string]string) error {
	q := url.Values{}
	for k, v := range data {
		q.Set(k, v)
	}
	err := db.Model(&Settings{}).Where("id = ?", c.ID).Update("value", q.Encode()).Error
	if err != nil {
		return err
	}
	return nil
}

func GetSetting(db *gorm.DB, name string) (Settings, error) {
	var settings Settings
	err := db.Where("name = ?", name).Last(&settings).Error
	if err != nil {
		return settings, err
	}
	return settings, nil
}
