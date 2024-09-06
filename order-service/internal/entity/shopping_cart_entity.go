package entity

type ShoppingCart struct {
	CartID int `json:"cart_id"`
	UserID int `json:"user_id"`
	Items  []ShoppingCartItem
}

type ShoppingCartItem struct {
	CartItemID int `json:"cart_item_id"`
	ProductID  int `json:"product_id"`
	Quantity   int `json:"quantity"`
}
