package domain

// Item representa un producto en la orden
type Item struct {
    SKU      string `json:"sku"`
    Quantity int    `json:"quantity"`
}

// NewItem crea un nuevo Item con validaciones
func NewItem(sku string, quantity int) (*Item, error) {
    if sku == "" {
        return nil, ErrInvalidSKU
    }
    
    if quantity <= 0 {
        return nil, ErrInvalidQuantity
    }
    
    return &Item{
        SKU:      sku,
        Quantity: quantity,
    }, nil
}

// Validate verifica que el Item sea válido
func (i *Item) Validate() error {
    if i.SKU == "" {
        return ErrInvalidSKU
    }
    
    if i.Quantity <= 0 {
        return ErrInvalidQuantity
    }
    
    return nil
}