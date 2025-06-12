package integration

import (
	"fmt"
	"net/url"
	"testing"
	"time"

	"github.com/miyamo2/qilin"
)

type Order struct {
	BeerID   string `json:"beer_id"`
	Quantity int    `json:"quantity"`
}

type OrderRequest struct {
	Orders []Order `json:"orders"`
}

type OrderResponse struct {
	Amount int `json:"amount"`
}

func OrderHandler(c qilin.ToolContext) error {
	var req OrderRequest
	if err := c.Bind(&req); err != nil {
		return err
	}
	var totalAmount int
	for _, order := range req.Orders {
		beer, ok := beers[order.BeerID]
		if !ok {
			return fmt.Errorf("beer %s does not exist", order.BeerID)
		}
		totalAmount += beer.Price * order.Quantity
	}
	return c.JSON(OrderResponse{Amount: totalAmount})
}

type GreetingArgs struct {
	Name string `json:"name"`
}

func GreetingPromptHandler(c qilin.PromptContext) error {
	var args GreetingArgs
	if err := c.Bind(&args); err != nil {
		return err
	}

	name := args.Name
	if name == "" {
		name = "World"
	}

	if err := c.String("system", "You are a helpful assistant."); err != nil {
		return err
	}

	return c.String("user", fmt.Sprintf("Hello, %s! How can I help you today?", name))
}

type Beer struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Price int    `json:"price"`
}

var beers = map[string]Beer{
	"1": {ID: "1", Name: "IPA", Price: 500},
	"2": {ID: "2", Name: "Stout", Price: 600},
	"3": {ID: "3", Name: "Lager", Price: 400},
	"4": {ID: "4", Name: "Pilsner", Price: 450},
}

func BeerListHandler(c qilin.ResourceContext) error {
	var beerList []Beer
	for _, beer := range beers {
		beerList = append(beerList, beer)
	}
	return c.JSON(beerList)
}

func BeerDetailHandler(c qilin.ResourceContext) error {
	beerID := c.Param("id")
	beer, ok := beers[beerID]
	if !ok {
		return fmt.Errorf("beer %s does not exist", beerID)
	}
	return c.JSON(beer)
}

func ResourceListHandler(c qilin.ResourceListContext) error {
	err := qilin.DefaultResourceListHandler(c)
	if err != nil {
		return err
	}
	for _, beer := range beers {
		uri, err := url.Parse(fmt.Sprintf("beer://detail/%s", beer.ID))
		if err != nil {
			return err
		}
		c.SetResource(uri.String(), qilin.Resource{
			URI:         (*qilin.ResourceURI)(uri),
			Name:        beer.Name,
			Description: fmt.Sprintf("Details of %s", beer.Name),
			MimeType:    "application/json",
		})
	}
	return nil
}

func ResourceChangeObserver(c qilin.ResourceChangeContext) {
	tick := time.Tick(10 * time.Second)
	for v := range tick {
		uri, err := url.Parse("beer://list")
		if err != nil {
			return
		}
		c.Publish(uri, v)
	}
	return
}

func ResourceListChangeObserver(c qilin.ResourceListChangeContext) {
	tick := time.Tick(10 * time.Second)
	for v := range tick {
		c.Publish(v)
	}
	return
}

func NewQilin(t *testing.T) *qilin.Qilin {
	t.Helper()

	q := qilin.New("beer_hall", qilin.WithVersion("1.0.0"))
	q.Tool("order", (*OrderRequest)(nil), OrderHandler)
	q.Prompt("greeting", GreetingPromptHandler,
		qilin.PromptWithDescription("A greeting prompt that welcomes users"),
		qilin.PromptWithArguments(
			qilin.NewPromptArgument("name", "The name of the person to greet", false),
		))
	q.Resource("beer_list", "beer://list", BeerListHandler)
	q.Resource("beer_detail", "beer://detail/{id}", BeerDetailHandler)
	q.ResourceList(ResourceListHandler)
	q.ResourceChangeObserver("beer://list", ResourceChangeObserver)
	q.ResourceListChangeObserver(ResourceListChangeObserver)

	return q
}
