package controllers

import (
	"github.com/revel/revel"

	"encoding/json"
	"net/http"
	"io/ioutil"
	"time"
	"regexp"
	"crypto/md5"
	"encoding/hex"
	"strings"
)

type CardEntity struct {
	Id			string	`json:"id"`
	Pan			string	`json:"pan"`
	ExpMonth	string	`json:"exp_month"`
	ExpYear		string	`json:"exp_year"`
}

type CardContainer []CardEntity

type Response struct {
	Success		bool	`json:"success"`
}

type Card struct {
	*revel.Controller
}

func (c Card) GetList() revel.Result {
	cardContainer := &CardContainer{}
	json.Unmarshal([]byte(c.Session["cards"]), cardContainer)
	return c.RenderJson(cardContainer)
}

func (c Card) Add() revel.Result {
	card := &CardEntity{}

	requestBody, _ := ioutil.ReadAll(c.Request.Body)
	err := json.Unmarshal(requestBody, card)
	if err != nil {
		c.Response.Status = http.StatusBadRequest
		return c.RenderText("Bad request")
	}

	if !validateCard(card) {
		return c.RenderJson(&Response{Success:false})
	}

	cardContainer := CardContainer{}
	json.Unmarshal([]byte(c.Session["cards"]), &cardContainer)
	cardContainerJson, _  := json.Marshal(append(cardContainer, *card))
	c.Session["cards"] = string(cardContainerJson)
	return c.RenderJson(card)
}

func (c Card) Delete(id string) revel.Result {
	cardContainer := CardContainer{}
	json.Unmarshal([]byte(c.Session["cards"]), &cardContainer)

	for index, card := range cardContainer {
		if card.Id == id {
			cardContainer = append(cardContainer[:index], cardContainer[index + 1:]...)
			cardContainerJson, _  := json.Marshal(cardContainer)
			c.Session["cards"] = string(cardContainerJson)
			return c.RenderJson(Response{Success:true})
		}
	}

	c.Response.Status = http.StatusNotFound
	return c.RenderText("Not found")
}

func validateCard(card *CardEntity) bool {
	date, err := time.Parse("01/2006", card.ExpMonth + "/" + card.ExpYear)
	if err != nil {
		return false
	}

	currentDate := time.Now()
	if (currentDate.Year() > date.Year()) ||
		(currentDate.Year() == date.Year() && currentDate.Month() > date.Month()) {
		return false
	}

	pan := strings.Replace(card.Pan, " ", "", -1)
	if regexp.MustCompile("[^0-9]").FindString(pan) != "" {
		return false
	}

	if len(pan) > 19 || len(pan) < 12 {
		return false
	}
	card.Pan = pan[len(pan) - 4:]

	hash := md5.Sum([]byte(time.Now().String()))
	card.Id = hex.EncodeToString(hash[:])

	return true
}

func initSession(c *revel.Controller) revel.Result {
	if _, initialized := c.Session["cards"]; !initialized {
		cardContainer := &CardContainer{}
		cardContainerJson, _  := json.Marshal(cardContainer)
		c.Session["cards"] = string(cardContainerJson)
	}
	return nil
}

func init() {
	revel.InterceptFunc(initSession, revel.BEFORE, &Card{})
}
