package dto

type UserDTO struct {
	Id       int32  `protobuf:"varint,1,opt,name=id,proto3" json:"id,omitempty"`
	Username string `protobuf:"bytes,2,opt,name=username,proto3" json:"username,omitempty"`
	Email    string `protobuf:"bytes,3,opt,name=email,proto3" json:"email,omitempty"`
	Role     string `protobuf:"bytes,4,opt,name=role,proto3" json:"role,omitempty"`
}

type OrderDTO struct {
	ID     int32   `json:"id,omitempty"`
	UserId int32   `json:"user_id"`
	ItemId int32   `json:"item_id"`
	Item   ItemDTO `json:"item"`
}

type ItemDTO struct {
	ID          int32  `json:"id,omitempty"`
	Name        string `json:"name,omitempty"`
	Price       int32  `json:"price,omitempty"`
	Description string `json:"description,omitempty"`
	Quantity    int32  `json:"quantity,omitempty"`
	ImageURL    string `json:"image_url"`
}
