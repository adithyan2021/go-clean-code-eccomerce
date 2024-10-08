package model

import (
	"fmt"
	"net/url"
	"time"
)

type Cart struct {
	Productid string  `json:"product_id"`
	Unit      float64 `json:"unit"`
	Userid    string  `json:"user_id"`
	To_delete bool    `json:"to_delete"`
}
type Usercartview struct {
	Productcategory string `json:"category"`
	Productname     string `json:"name"`
	Unit            int    `json:"unit"`
	Productamount   int    `json:"amount"`
	Productstatus   bool   `json:"status"`
	Vendorname      string `json:"vendorName"`
}
type UserWishview struct {
	Productcategory string `json:"category"`
	Productname     string `json:"name"`

	Productamount int    `json:"amount"`
	Productstatus bool   `json:"status"`
	Vendorname    string `json:"vendorName"`
}

type Wishlist struct {
	Productid string `json:"product_id"`

	Userid string
}
type Coupon struct {
	Code      string `json:"code"`
	Expiry    string `json:"expiry"`
	Minamount int    `json:"min_amount"`
	Amount    int    `json:"amount"`
	Maxamount int    `json:"max_amount"`
}

func ValidateExpiry(expiry string) error {
	const dateFormat = "2006-01-02" // Layout for YYYY-MM-DD format
	_, err := time.Parse(dateFormat, expiry)
	if err != nil {
		return fmt.Errorf("expiry date should be in YYYY-MM-DD format")
	}
	return nil
}
func (c *Coupon) Valid() (err url.Values) {
	err = url.Values{}

	if c.Code == "" {
		err.Add("Code error", "enter code name")
	}
	if c.Minamount < 0 {
		err.Add("Minamount error", "enter Minamount")
	}
	if c.Maxamount < 0 {
		err.Add("Maxamount error", "enter Maxamount")
	}
	if c.Amount < 0 || c.Amount > 70 {
		err.Add("Amount error", "enter Amount")
	}
	if c.Expiry == "" {
		err.Add("Expiry error", "enter Expiry ")
	} else {
		// Validate the expiry date format
		if validationErr := ValidateExpiry(c.Expiry); validationErr != nil {
			err.Add("Expiry error", validationErr.Error())
		}
	}

	return err
}

type CouponRes struct {
	Cid         string `json:"cid"`
	Code        string `json:"code"`
	Expiry      string `json:"expiry"`
	Minamount   int    `json:"min_amount"`
	Amount      int    `json:"amount"`
	CurrentDate string `json:"current_date"`
	Is_expired  bool   `json:"is_expired"`
	Is_eligible bool   `json:"is_eligible"`
	Used        bool   `json:"used"`
	Maxamount   int    `json:"max_amount"`
	Present     bool
}
type Cartresponse struct {
	Cid      string `json:"cid"`
	Pid      string `json:"pid"`
	Usid     string `json:"usid"`
	Amount   int    `json:"amount"`
	Unit     int    `json:"unit"`
	Discount int    `json:"discount"`
	P_Units  int    `json:"p_units"`
	P_Name   string `json:"p_name"`
}
type CartresponseData struct {
	Data []Cartresponse
	Err  error
}
type FirstAddOrder struct {
	Data     CartresponseData
	TAmount  int
	CData    CouponRes
	Notvalid bool
}

type RZpayment struct {
	Id              string
	Amt             float64
	Token           string
	Order_ID        string
	CartData        CartresponseData
	User_id         string
	WalletDeduction float64
	Oid             string
	Cid             string
}
type Order struct {
	Cartid       string `json:"cart_id"`
	Couponid     string `json:"coupon_id"`
	Type         string `json:"cod"`
	Returnstatus bool   `json:"returnstatus"`
	Aid          string `json:"aid"`
}

func (u *Order) Valid() url.Values {
	err := url.Values{}

	if u.Aid == "" {
		err.Add("Address ", "please ADD Address")
		return err
	}
	if u.Type == "" {
		err.Add("Payment Type  ", "please ADD Payment Type")
		return err
	}
	// if u.Aid == "" {
	// 	err.Add("Aid ", "no address")
	// 	return err
	// }
	return url.Values{}

}

type CheckOut struct {
	Couponid     string `json:"coupon_id"`
	Type         string `json:"cod"`
	Returnstatus bool   `json:"returnstatus"`
	Aid          string `json:"aid"`
	Wallet       bool   `json:"w_amt"`
}

func (u *CheckOut) Valid() (err url.Values, Coupon bool) {
	err = url.Values{}
	Coupon = false
	if u.Aid == "" {
		err.Add("Address ", "please ADD Address")

	}
	if !(u.Type == "ONLINE" || u.Type == "COD") {
		err.Add("Payment Type  ", "please ADD Payment Type")

	}
	// if u.Aid == "" {
	// 	err.Add("Aid ", "no address")
	// 	return err
	// }
	return err, Coupon

}

type Placeorderlist struct {
	Data string
	Err  error
}
type InsertOrder struct {
	Usid       string
	Amount     int
	Discount   int
	CouponAmt  float64
	WalletAmt  float64
	PayableAmt float64
	PayType    string
	Aid        string
	Status     string
	CouponId   string
}
type PaymentInsert struct {
	OrderId string
	Usid    string
	Amount  float64
	Status  string
	Type    string
}

func (u *CouponRes) Valid() (err url.Values) {
	err = url.Values{}

	if !u.Is_eligible {
		fmt.Println("in check 1!!!")
		err.Add("Amount ", "Total amount is less")

	}
	if u.Is_expired {
		fmt.Println("in check 2!!!")
		err.Add("Expired ", "coupon is expired")

	}
	if u.Used {
		fmt.Println("in check 3!!!")
		err.Add("Used ", "coupon is already used")

	}
	// if u.Aid == "" {
	// 	err.Add("Aid ", "no address")
	// 	return err
	// }
	return err

}
