package user

import (
	"context"
	"math"
	"net/url"
	"sync"

	"fmt"
	"github.com/razorpay/razorpay-go"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	services "myproject/pkg/client"
	"myproject/pkg/config"
	"myproject/pkg/model"
	"regexp"
	"strings"
)

type Service interface {
	Register(ctx context.Context, request model.UserRegisterRequest) error
	Login(ctx context.Context, request model.UserLoginRequest) error
	//Product listing
	Listing(ctx context.Context) ([]model.ProductListingUsers, error)
	BrandListing(ctx context.Context, category string) ([]model.ProductListingUsers, error)
	CategoryListing(ctx context.Context, category string) ([]model.ProductListingUsers, error)
	ListingSingle(ctx context.Context, id string) ([]model.ProductListDetailed, error)
	LatestListing(ctx context.Context) ([]model.ProductListingUsers, error)
	PhighListing(ctx context.Context) ([]model.ProductListingUsers, error)
	PlowListing(ctx context.Context) ([]model.ProductListingUsers, error)
	InAZListing(ctx context.Context) ([]model.ProductListingUsers, error)
	InZAListing(ctx context.Context) ([]model.ProductListingUsers, error)
	BestSellingListingProductCategory(ctx context.Context, category string) ([]model.ProductListingUsers, error)
	BestSellingListingProductBrand(ctx context.Context, category string) ([]model.ProductListingUsers, error)
	BestSellingListingProduct(ctx context.Context) ([]model.ProductListingUsers, error)

	BestSellingListingCategory(ctx context.Context) ([]string, error)
	BestSellingListingBrand(ctx context.Context) ([]string, error)
	////
	OtpLogin(ctx context.Context, request model.UserOtp) error
	UpdateUser(ctx context.Context, updatedData model.UserRegisterRequest) error

	////cart and wish
	AddTocart(ctx context.Context, request model.Cart, username string) error
	AddToWish(ctx context.Context, request model.Wishlist) error
	UpdateToCart(ctx context.Context, request model.Cart, username string) error
	///
	AddAddress(ctx context.Context, request model.Address, username string) error
	AddToorder(ctx context.Context, request model.Order) (model.RZpayment, error)
	Listcart(ctx context.Context, id string) ([]model.Usercartview, error)
	ListWish(ctx context.Context, id string) ([]model.UserWishview, error)
	ListAddress(ctx context.Context, username string) ([]model.Address, error)
	ActiveListing(ctx context.Context) ([]model.Coupon, error)
	AddToCheck(ctx context.Context, request model.CheckOut, username string) (model.RZpayment, error, url.Values)
	PaymentSuccess(ctx context.Context, rz model.RZpayment, username string) error
	PaymentFailed(ctx context.Context, rz model.RZpayment, username string) error
	//list orders
	ListAllOrders(ctx context.Context, username string) ([]model.ListAllOrdersUsers, error)
	ListReturnedOrders(ctx context.Context, username string, status string) ([]model.ListAllOrdersUsers, error)
	ListFailedOrders(ctx context.Context, username string) ([]model.ListAllOrdersUsers, error)
	ListCompletedOrders(ctx context.Context, username string) ([]model.ListAllOrdersUsers, error)
	ListPendingOrders(ctx context.Context, username string) ([]model.ListAllOrdersUsers, error)
	ReturnItem(ctx context.Context, request model.ReturnOrderPostForUser, username string) error

	ListAllTransactions(ctx context.Context, username string) ([]model.UserTransactions, error)
	ListTypeTransactions(ctx context.Context, username string, ty string) ([]model.UserTransactions, error)

	/// list main orders
	ListMainOrders(ctx context.Context, username string) ([]model.ListingUserMainOrders, error)
	CancelMainOrders(ctx context.Context, username string, orderUid string) error
	GetMainOrders(ctx context.Context, username string, orderUid string) ([]model.ListAllOrdersUsers, error)
	VerifyOtp(ctx context.Context, email string)
}
type service struct {
	repo     Repository
	Config   config.Config
	services services.Services
}

func NewService(repo Repository, services services.Services) Service {
	return &service{
		repo:     repo,
		services: services,
	}
}
func (s *service) VerifyOtp(ctx context.Context, email string) {
	s.repo.VerifyOtp(ctx, email)

}
func (s *service) GetMainOrders(ctx context.Context, username string, orderUid string) ([]model.ListAllOrdersUsers, error) {

	id := s.repo.Getid(ctx, username)
	/// check if order id is valid

	orders, err := s.repo.PrintingUserSingleMainOrderCollection(ctx, id, orderUid)
	fmt.Println(orders)
	if err != nil {
		return []model.ListAllOrdersUsers{}, fmt.Errorf("error in retriving data")
	}

	return orders, nil

}
func (s *service) CancelMainOrders(ctx context.Context, username string, orderUid string) error {

	id := s.repo.Getid(ctx, username)
	/// check if order id is valid

	orders, err := s.repo.PrintingUserSingleMainOrder(ctx, id, orderUid)
	fmt.Println(orders)
	if err != nil {
		return fmt.Errorf("error in retriving data")
	}
	for _, v := range orders {
		err := s.ReturnItem(ctx, v, username)
		if err != nil {
			return fmt.Errorf("error in cancelling ", err)
		}

	}
	er := s.repo.ChangeOrderStatus(ctx, orderUid)
	fmt.Println(er)
	if er != nil {
		return fmt.Errorf("error in updating ")
	}

	return nil

}
func (s *service) ListMainOrders(ctx context.Context, username string) ([]model.ListingUserMainOrders, error) {

	id := s.repo.Getid(ctx, username)
	orders, err := s.repo.PrintingUserMainOrder(ctx, id)
	if err != nil {
		return []model.ListingUserMainOrders{}, fmt.Errorf("error in retriving data")
	}
	return orders, nil

}

// /transactions
func (s *service) ListAllTransactions(ctx context.Context, username string) ([]model.UserTransactions, error) {
	id := s.repo.Getid(ctx, username)

	transactions, err := s.repo.ListAllTransactions(ctx, id)
	if err != nil {

		return []model.UserTransactions{}, fmt.Errorf("there is error in fetching transactions")
	}

	return transactions, nil

}
func (s *service) ListTypeTransactions(ctx context.Context, username string, ty string) ([]model.UserTransactions, error) {
	id := s.repo.Getid(ctx, username)

	transactions, err := s.repo.ListTypeTransactions(ctx, id, ty)
	if err != nil {

		return []model.UserTransactions{}, fmt.Errorf("there is error in fetching transactions")
	}

	return transactions, nil

}

// ///
func (s *service) ReturnItem(ctx context.Context, request model.ReturnOrderPostForUser, username string) error {
	id := s.repo.Getid(ctx, username)
	fmt.Println("inside the ReturnItem ", id)
	p, err := s.repo.GetSingleItem(ctx, id, request.Oid)
	fmt.Println("this is the ppppp", p)

	if err != nil {
		return fmt.Errorf("entered is wrong id", err)
	}
	fmt.Println("this is the single order ", p)
	switch p.Returned {
	case "Returned":
		return fmt.Errorf("this item is already returned")

	case "Cancelled":
		return fmt.Errorf("this item is  Cancelled")

	}

	if p.Delivery == "Completed" && request.Type == "Cancelled" {
		return fmt.Errorf("can not cancel returned item")
	}
	if p.Delivery == "Pending" && request.Type == "Returned" {
		return fmt.Errorf("can not return before delivering")
	}

	switch p.Status {
	case "Failed":
		return fmt.Errorf("this item is payment failed")

	case "Cancelled":
		return fmt.Errorf("this item is payment Cancelled")

	}

	var w sync.WaitGroup

	Err := make(chan error, 1)
	Err2 := make(chan error, 1)
	Err3 := make(chan error, 1)
	Err4 := make(chan error, 1)
	//Err5 := make(chan error, 1)

	w.Add(4)
	go func() {
		defer w.Done()
		err := s.services.SendOrderReturnConfirmationEmailUser(p.Name, p.Amount, p.Unit, username)
		Err <- err
	}()
	go func() {
		defer w.Done()
		err := s.repo.IncreaseStock(ctx, p.Pid, p.Unit)
		Err2 <- err
	}()
	go func() {
		defer w.Done()
		err := s.repo.UpdateOiStatus(ctx, request.Oid, request.Type)
		Err3 <- err
	}()
	go func() {
		defer w.Done()
		var err error
		var wallet_id string
		if p.Status == "Completed" {
			fmt.Println("in 1st if")
			// value := []interface{}{p.Amount, id, "Credit"}
			fmt.Println("this is the serviceee !@!!#!## ---", p.Moid)
			cmp_r_amt, _ := s.repo.GetcpAmtRefund(ctx, p.Moid)
			fmt.Println("this is the amount !!!", p.Amount, "c ppp amtt", cmp_r_amt)
			var f float64
			var upcmCheck string
			if p.Amount*float64(p.Unit) > float64(cmp_r_amt) {
				f = p.Amount*float64(p.Unit) - float64(cmp_r_amt)
				upcmCheck = "normal"
			} else {
				f = p.Amount * float64(p.Unit)
				upcmCheck = "notnormal"
			}
			wallet_id, err = s.repo.CreditWallet(ctx, id, f)
			fmt.Println("wallletttt---", wallet_id)

			if wallet_id != "" {
				value := []interface{}{f, wallet_id, "Credit", id, p.Moid}
				fmt.Println("pppp,", p.Moid, "--", value)
				er := s.repo.UpdateWalletTransaction(ctx, value)
				if er != nil {
					fmt.Println("there is erorrrr in wallet transaction")
					err = er

				} else {
					///updateing the mo statussssss
					if upcmCheck == "normal" {
						s.repo.ChangeCouponRefundStatus(ctx, p.Moid)
						fmt.Println(" normalll")
					}

				}
				//ers

			}
		} else {
			fmt.Println("in 1st else")
			wallet_id = ""
			err = nil
		}
		Err4 <- err

	}()

	go func() {
		w.Wait()
		close(Err)
		close(Err2)
		close(Err3)
		close(Err4)

	}()

	if err := <-Err; err != nil {
		return fmt.Errorf("failed to send order  return  email: %w", err)
	}
	if err := <-Err2; err != nil {
		return fmt.Errorf("failed to update unit: %w", err)
	}
	if err := <-Err3; err != nil {
		return fmt.Errorf("failed to update to refund status: %w", err)
	}
	if err := <-Err4; err != nil {
		return fmt.Errorf("failed to update to refund status: %w", err)
	}

	return nil
}

func (s *service) ListAllOrders(ctx context.Context, username string) ([]model.ListAllOrdersUsers, error) {
	id := s.repo.Getid(ctx, username)
	fmt.Println("inside the ListAllOrders ", id)
	orders, err := s.repo.ListAllOrders(ctx, id)
	if err != nil {
		return []model.ListAllOrdersUsers{}, fmt.Errorf("this is the error for listing all orders", err)
	}

	return orders, nil
}
func (s *service) ListReturnedOrders(ctx context.Context, username string, status string) ([]model.ListAllOrdersUsers, error) {
	id := s.repo.Getid(ctx, username)
	fmt.Println("inside the ListAllOrders ", id)
	orders, err := s.repo.ListReturnedOrders(ctx, id, status)
	if err != nil {
		return []model.ListAllOrdersUsers{}, fmt.Errorf("this is the error for listing all orders", err)
	}

	return orders, nil
}
func (s *service) ListFailedOrders(ctx context.Context, username string) ([]model.ListAllOrdersUsers, error) {
	id := s.repo.Getid(ctx, username)
	fmt.Println("inside the ListAllOrdersUsers ", id)
	orders, err := s.repo.ListFailedOrders(ctx, id)
	if err != nil {
		return []model.ListAllOrdersUsers{}, fmt.Errorf("this is the error for listing all orders", err)
	}

	return orders, nil
}
func (s *service) ListCompletedOrders(ctx context.Context, username string) ([]model.ListAllOrdersUsers, error) {
	id := s.repo.Getid(ctx, username)
	fmt.Println("inside the ListAllOrdersUsers ", id)
	orders, err := s.repo.ListCompletedOrders(ctx, id)
	if err != nil {
		return []model.ListAllOrdersUsers{}, fmt.Errorf("this is the error for listing all orders", err)
	}

	return orders, nil
}
func (s *service) ListPendingOrders(ctx context.Context, username string) ([]model.ListAllOrdersUsers, error) {
	id := s.repo.Getid(ctx, username)
	fmt.Println("inside the ListAllOrdersUsers ", id)
	orders, err := s.repo.ListPendingOrders(ctx, id)
	if err != nil {
		return []model.ListAllOrdersUsers{}, fmt.Errorf("this is the error for listing all orders", err)
	}

	return orders, nil
}

func (s *service) ListAddress(ctx context.Context, username string) ([]model.Address, error) {
	id := s.repo.Getid(ctx, username)
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		return s.repo.ListAddress(ctx, id)
	}

}
func (s *service) ActiveListing(ctx context.Context) ([]model.Coupon, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		return s.repo.ActiveListing(ctx)
	}

}
func (s *service) AddTocart(ctx context.Context, request model.Cart, username string) error {

	id := s.repo.Getid(ctx, username)
	if request.Productid == "" || request.Unit == 0 {
		fmt.Println("this is in the service error value missing")
		return fmt.Errorf("missing values")
	}
	if request.Unit < 0 {
		return fmt.Errorf("enter valid unit")

	}

	products, err := s.repo.ListingByid(ctx, request.Productid)
	//check if product already exist
	check, err := s.repo.ItemExistsInCart(ctx, request.Productid, id)
	if err != nil {
		fmt.Println("failed to get product:", err)
		return fmt.Errorf("failed to check product exist or not: %w", err)
	}
	if check {
		return fmt.Errorf("Product is already added ")

	}

	if err != nil {
		fmt.Println("failed to get product:", err)
		return fmt.Errorf("failed to get product: %w", err)
	}

	if len(products) == 0 {
		fmt.Println("product not found")
		return fmt.Errorf("product not found")
	}

	product := products[0]
	fmt.Println("this is in cart ", product.Unit)
	if product.Unit == 0 {

		return fmt.Errorf("not enough units available")

	}
	if product.Unit < float64(request.Unit) {
		fmt.Println("not enough units available")
		return fmt.Errorf("not enough units available")
	}
	request.Userid = id
	// Add to cart
	return s.repo.AddTocart(ctx, request)
}
func (s *service) UpdateToCart(ctx context.Context, request model.Cart, username string) error {

	id := s.repo.Getid(ctx, username)
	if request.Productid == "" {
		fmt.Println("this is in the service error value missing")
		return fmt.Errorf("missing values S")
	}
	if request.To_delete {
		fmt.Println("this is the product id", request.Productid)
		err := s.repo.DeleteSingleCart(ctx, id, request.Productid)
		if err != nil {
			return fmt.Errorf("error in deleting cart try again")
		} else {
			return nil
		}

	}
	products, err := s.repo.ListingByid(ctx, request.Productid)

	if err != nil {
		fmt.Println("failed to get product:", err)
		return fmt.Errorf("failed to get product: %w", err)
	}

	if len(products) == 0 {
		fmt.Println("product not found")
		return fmt.Errorf("product not found")
	}
	carts, err := s.repo.GetSpecificCart(ctx, id, request.Productid)
	if err != nil {
		fmt.Println("failed to get cart:", err)
		return fmt.Errorf("failed to get cart: %w", err)
	}

	product := products[0]
	fmt.Println("this is in cart ", product.Unit)
	if product.Unit == 0 {

		return fmt.Errorf("not enough units available")

	}
	if product.Unit < float64(request.Unit)+carts.Unit {
		fmt.Println("not enough units available")
		return fmt.Errorf("not enough units available")
	}
	if request.Unit < 0 {
		if math.Abs(request.Unit) > carts.Unit {
			return fmt.Errorf("enter valid units")
		}
	}
	request.Userid = id
	request.Unit = float64(carts.Unit) + float64(request.Unit)
	// Add to cart
	return s.repo.UpdateToCart(ctx, request)
}
func (s *service) AddToWish(ctx context.Context, request model.Wishlist) error {
	if request.Productid == "" {
		fmt.Println("this is in the service error value missing")
		return fmt.Errorf("missing values")
	}
	id := s.repo.Getid(ctx, request.Userid)
	request.Userid = id
	// _, err := s.repo.ListingByid(ctx, request.Productid)
	// if err != nil {
	// 	fmt.Println("failed to get product:", err)
	// 	return fmt.Errorf("failed to get product", err)
	// }

	return s.repo.AddToWish(ctx, request)
}
func (s *service) AddAddress(ctx context.Context, request model.Address, username string) error {
	// if request.Address == "" {
	// 	fmt.Println("this is in the service error value missing")
	// 	return fmt.Errorf("missing values")
	// }
	id := s.repo.Getid(ctx, username)
	fmt.Println("this is my idddddddd!!!!", id)

	return s.repo.AddAddress(ctx, request, id)
}
func (s *service) AddToorder(ctx context.Context, request model.Order) (model.RZpayment, error) {
	//ctx context.Context, request model.Order
	fmt.Println("inside the service Addto order ")
	username := ctx.Value("username").(string)
	fmt.Println("inside the cart list ", username)
	fiData, _ := s.repo.GetorderDetails(ctx, request)
	fmt.Println("thisss is the daaa ", fiData.Data.Data, "and this is amount ", fiData.TAmount)
	if !fiData.Notvalid {
		return model.RZpayment{}, fmt.Errorf("not a valid coupon")
	}
	var k model.RZpayment
	k = s.PayGateway(ctx, fiData.TAmount)
	fmt.Println("this is kkkkkkkk!!!!!", k)
	// Pid, err := s.repo.AddToPayment(ctx, request, fiData, status, username)
	// oid, err := s.repo.AddToOrderDetails(ctx, request, fiData, status, username, Pid)
	//fmt.Println("this is the status ", status)
	return k, nil
}
func (s *service) AddToCheck(ctx context.Context, request model.CheckOut, username string) (model.RZpayment, error, url.Values) {
	id := s.repo.Getid(ctx, username)
	Paymentstatus := "Pending"
	fmt.Println("inside the service corrected Addto order ", id)
	cartDataChan := make(chan model.CartresponseData)
	amountChan := make(chan int)
	discountChan := make(chan int)
	go func() {
		data, err := s.repo.GetcartRes(ctx, id)
		cartDataChan <- model.CartresponseData{Data: data, Err: err}
		close(cartDataChan)

	}()
	go func() {
		amount := s.repo.GetcartAmt(ctx, id)
		amountChan <- amount
		close(amountChan)

	}()
	go func() {
		amount := s.repo.GetcartDis(ctx, id)
		discountChan <- amount
		close(discountChan)

	}()

	cartData := <-cartDataChan

	amount := <-amountChan
	discount := <-discountChan
	checker := func(cartData model.CartresponseData) url.Values {
		err := url.Values{}
		for _, v := range cartData.Data {
			fmt.Println("Processing item:", v.P_Name, v.P_Units, v.Unit)
			if v.P_Units < v.Unit {
				erS := v.P_Name + " this is product in low in stock"
				err.Add(v.P_Name, erS)
			}

		}
		return err

	}
	stErr := checker(cartData)
	fmt.Println("Stock error", stErr)
	if len(stErr) > 0 {
		fmt.Println("in this errr")
		return model.RZpayment{}, nil, stErr
	}

	cidc := request.Couponid
	cid, err := s.repo.GetCoupnExist(ctx, cidc)

	// if err!=nil {
	//   return model.RZpayment{},fmt.Errorf("Error in retriving the coupon"),nil
	// }

	PayType := request.Type
	fmt.Println("klkl", PayType)
	couponAmt := 0
	var maxCAmt int
	var coupon = model.CouponRes{}
	k, err := s.repo.GetCartExist(ctx, id)
	fmt.Println("checking exist or notttttt!!!!!!!", k)
	if k == "" {
		fmt.Println("ifffff   checking exist in ifffff notttttt!!!!!!!")
		return model.RZpayment{}, fmt.Errorf("no order exist for the user: %w"), nil

	}
	if cid != "" {
		fmt.Println("this is cccccccc-----ccccc1111", cid)
		coupon = s.repo.GetCoupon(ctx, cid, amount)
		checkCo, _ := s.repo.CheckCouponExist(ctx, cid, id)
		fmt.Println("checkCo !!! workinggg !!!!", checkCo)
		if checkCo {
			fmt.Println("checkCo !!! workinggg")
			return model.RZpayment{}, fmt.Errorf("invalid", "coupon already added"), nil
		}

		fmt.Println("this is cccccccc-----ccccc", checkCo)
		fmt.Println("this is cccccccc-----22", coupon)
		errValues := coupon.Valid()
		if len(errValues) > 0 {
			// return fmt.Errorf(map[string]interface{}{"invalid": errValues})
			return model.RZpayment{}, fmt.Errorf("invalid", errValues), nil
		}
		couponAmt = coupon.Amount

		maxCAmt = coupon.Maxamount
		fmt.Println("this is coupon!!!!!", coupon, "this is checks ", coupon.Is_expired, "this is eligible ", coupon.Is_eligible)

	}
	wamt := request.Wallet
	var w_amt float32
	if wamt == true {
		w_amt = s.repo.GetWallAmt(ctx, id, amount)
		fmt.Println("this is walett!!!!!", w_amt)

	} else {
		w_amt = 0.0
	}
	if request.Type == "COD" && amount < 1000 {
		return model.RZpayment{}, fmt.Errorf("in cash on delivery should be greater than 1000"), nil
	}

	Camt := (float64(couponAmt) / 100.0) * float64(amount)
	fmt.Println("this is the comparison of coupon amt --- ", Camt, " --max amt - ", maxCAmt)
	if maxCAmt < int(Camt) {

		return model.RZpayment{}, fmt.Errorf("this is more than the limit of the coupon"), nil
	}
	fmt.Println("this is the calculation vart!!", Camt, couponAmt, amount, w_amt, PayType, request.Aid, Paymentstatus)
	var newAmount float64
	walletDeduction := 0.0
	if float64(w_amt) < (float64(amount) - Camt) {

		newAmount = float64(amount) - Camt - float64(w_amt)
		walletDeduction = float64(w_amt)
		fmt.Println("1 this is the calculated amount!", newAmount, "Previous Amount", amount)
	} else {
		newAmount = 0
		request.Type = "Wallet"
		Paymentstatus = "Completed"
		walletDeduction = float64(amount) - Camt
		fmt.Println("2 this is the calculated amount!", newAmount, "Previous Amount", amount)

	}
	fmt.Println("this is the end data from AddToCheck", cartData, "dvdsvdsvdsv!!", amount, "this is the new amount", newAmount, "wallet deduction!!", walletDeduction)
	order := model.InsertOrder{
		Usid:       id,
		Amount:     amount,
		Discount:   discount,
		CouponAmt:  Camt,
		WalletAmt:  walletDeduction,
		PayableAmt: newAmount,
		PayType:    PayType,
		Aid:        request.Aid,
		Status:     Paymentstatus,
		CouponId:   cid,
	}
	fmt.Println("this is cid in  checkout ", cid)
	OrderID, uuid, err := s.repo.CreateOrder(ctx, order)
	if err != nil {
		return model.RZpayment{}, fmt.Errorf("failed to create order: %w", err), nil
	}

	fmt.Println("Order created with ID:!!!!", OrderID, uuid)

	/// seng email
	/// sending Email
	var w sync.WaitGroup
	Err := make(chan error, 1) // Buffered channel to avoid deadlock

	w.Add(1)
	go func() {
		defer w.Done()
		err := s.services.SendOrderConfirmationEmail(uuid, newAmount, username)
		Err <- err // Send the error or nil to the channel
	}()

	w.Wait()
	close(Err)

	if err := <-Err; err != nil {
		return model.RZpayment{}, fmt.Errorf("failed to send order confirmation email: %w", err), nil
	}
	ty := request.Type
	paySt := model.PaymentInsert{
		OrderId: OrderID,
		Usid:    id,
		Amount:  newAmount,
		Status:  Paymentstatus,
		Type:    ty,
	}
	PaymentId, err := s.repo.MakePayment(ctx, paySt)
	if err != nil {
		return model.RZpayment{}, fmt.Errorf("failed to create order: %w", err), nil
	}

	err = s.repo.AddOrderItems(ctx, cartData, OrderID, id, PaymentId)
	if err != nil {
		return model.RZpayment{}, fmt.Errorf("failed to create order: %w", err), nil
	}

	fmt.Println("this the payment id!!", PaymentId)

	if ty != "ONLINE" || newAmount == 0 {
		j := make(chan error, 1)
		p := make(chan int)
		var wg sync.WaitGroup
		wg.Add(2)
		go func() {
			defer wg.Done()
			jf := s.PostCheckout(ctx, PaymentId, OrderID, cartData, id, walletDeduction)
			p <- jf
		}()
		go func() {
			defer wg.Done()
			if cid != "" {
				fmt.Println("this is the couponnnnn!!!!!", cid)
				err := s.repo.UpdateUsestatusCoupon(ctx, cid, id)
				if err != nil {
					j <- err
				} else {
					j <- nil
				}

			}
		}()
		go func() {
			wg.Wait()
			close(p)
			close(j) // Close the channel after all goroutines are done
		}()
		kl := <-p
		if kl != 1 {
			return model.RZpayment{}, fmt.Errorf("failed to create order: %w", err), nil

		}
		Lrr := <-j
		if Lrr != nil {
			return model.RZpayment{}, fmt.Errorf("failed to create order: %w", err), nil
		}

	} else {

		var rz model.RZpayment
		rz.Amt = newAmount
		rz.Id = PaymentId
		rz.Order_ID = uuid
		rz.CartData = cartData
		rz.WalletDeduction = walletDeduction
		rz.User_id = id
		rz.Oid = OrderID
		rz.Cid = cid

		return rz, nil, nil

	}

	return model.RZpayment{}, nil, nil
}

func (s *service) PaymentFailed(ctx context.Context, rz model.RZpayment, username string) error {
	fmt.Println("inside PaymentFailed")
	//id := s.repo.Getid(ctx, username)
	OrUpstat := make(chan error, 1)
	PayUpstat := make(chan error, 1)
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		UO := s.repo.UpdateOrderStatus(ctx, rz.Order_ID, "Failed")

		if UO != nil {
			OrUpstat <- UO
		} else {
			OrUpstat <- nil
		}
		//close(OrUpstat)
	}()

	go func() {
		defer wg.Done()
		UP := s.repo.UpdatePaymentStatus(ctx, rz.Id, "Failed")

		if UP != nil {
			PayUpstat <- UP
		} else {
			PayUpstat <- nil
		}

		//close(PayUpstat)
	}()
	go func() {
		wg.Wait()
		close(OrUpstat)
		close(PayUpstat)
	}()
	Lrr1 := <-OrUpstat
	if Lrr1 != nil {
		return fmt.Errorf("failed to post order:")
	}
	fmt.Println("in ! 1 check")
	Lrr2 := <-PayUpstat
	if Lrr2 != nil {
		return fmt.Errorf("failed to post order:")
	}

	return nil
}
func (s *service) PaymentSuccess(ctx context.Context, rz model.RZpayment, username string) error {
	fmt.Println("inside PaymentSuccess")

	id := s.repo.Getid(ctx, username)

	OrUpstat := make(chan error, 1)
	PayUpstat := make(chan error, 1)
	PostCheck := make(chan int)
	CoUpstat := make(chan error, 1)

	var wg sync.WaitGroup

	wg.Add(4)

	go func() {
		defer wg.Done()
		UO := s.repo.UpdateOrderStatus(ctx, rz.Order_ID, "Completed")

		if UO != nil {
			OrUpstat <- UO
		} else {
			OrUpstat <- nil
		}
		//close(OrUpstat)
	}()

	go func() {
		defer wg.Done()
		UP := s.repo.UpdatePaymentStatus(ctx, rz.Id, "Completed")

		if UP != nil {
			PayUpstat <- UP
		} else {
			PayUpstat <- nil
		}

		//close(PayUpstat)
	}()
	go func() {
		defer wg.Done()
		if rz.Cid != "" {
			fmt.Println("this is the couponnnnn!!!!!", rz.Cid)
			err := s.repo.UpdateUsestatusCoupon(ctx, rz.Cid, id)
			if err != nil {
				CoUpstat <- err
			} else {
				CoUpstat <- nil
			}

		}
	}()

	go func() {
		defer wg.Done()
		jf := s.PostCheckout(ctx, rz.Id, rz.Oid, rz.CartData, id, rz.WalletDeduction)
		PostCheck <- jf
		//close(PostCheck)
	}()

	go func() {
		wg.Wait()
		close(OrUpstat)
		close(PayUpstat) // Close the channel after all goroutines are done
		close(PostCheck)
		close(CoUpstat)
	}()

	Lrr1 := <-OrUpstat
	if Lrr1 != nil {
		return fmt.Errorf("failed to post order:")
	}
	fmt.Println("in ! 1 check")
	Lrr2 := <-PayUpstat
	if Lrr2 != nil {
		return fmt.Errorf("failed to post order:")
	}
	fmt.Println("in ! 2 check")
	kl := <-PostCheck
	if kl != 1 {
		return fmt.Errorf("failed to post order:")

	}
	fmt.Println("in ! 2 check")

	return nil
}

func (s *service) PostCheckout(ctx context.Context, PaymentId string, OrderID string, cartData model.CartresponseData, id string, walletDeduction float64) int {
	fmt.Println("this is in PostCheckout ", PaymentId, "#", OrderID, "#", cartData)
	fmt.Println("this is in PostCheckout 222!!", id, "#", walletDeduction)
	for _, v := range cartData.Data {
		quantity := v.Unit
		id := v.Pid
		value := []interface{}{quantity, id}

		err := s.repo.UpdateStock(ctx, value)
		if err != nil {
			fmt.Errorf("failed to create order: %w", err)
		}

	}
	value := []interface{}{walletDeduction, id}
	wallet_id, err := s.repo.UpdateWallet(ctx, value)
	fmt.Println("this is wallet id $$$$!!!$$$", wallet_id)
	if err != nil {
		fmt.Errorf("failed to create order: %w", err)
	}
	if walletDeduction != 0 {
		value := []interface{}{walletDeduction, wallet_id, "Debit", id}
		err := s.repo.UpdateWalletTransaction(ctx, value)
		if err != nil {
			fmt.Errorf("failed to create order: %w", err)
		}
	}
	err = s.repo.DeleteCart(ctx, id)
	if err != nil {
		fmt.Errorf("failed to create order: %w", err)
	}
	fmt.Println("completed 3")
	return 1

}
func (s *service) PayGateway(ctx context.Context, amt int) model.RZpayment {

	Razor_ID := "rzp_test_mRydipg2bgDZmQ"
	Razor_Secret := "a2oY1G5RYIQh9gH04KWATpnx"
	fmt.Println(" RZ keyyyyy ", Razor_ID, "  RZ ZSCERET ", Razor_Secret)
	razorpayClient := razorpay.NewClient(Razor_ID, Razor_Secret)
	data := map[string]interface{}{
		"amount":   amt,
		"currency": "INR",
		"receipt":  "some_receipt_id",
	}
	body, _ := razorpayClient.Order.Create(data, nil)

	value := body["id"]
	fmt.Println(" in bodyyyyyyy id!!!!!", value)
	str := value.(string)

	var p = model.RZpayment{
		Id: str,
		// Amt: amt,
	}
	fmt.Println("ggg passs on RZPAy", p)

	return p
}

type PageVariable struct {
	AppointmentID string
}

func (s *service) Register(ctx context.Context, request model.UserRegisterRequest) error {
	var err error
	if request.FirstName == "" || request.Email == "" || request.Password == "" || request.Phone == "" {
		fmt.Println("this is in the service error value missing")
		err = fmt.Errorf("missing values")
		return err
	}
	if !isValidEmail(request.Email) {
		fmt.Println("this is in the service error invalid email")
		err = fmt.Errorf("invalid email")
		return err
	}
	if !isValidPhoneNumber(request.Phone) {
		fmt.Println("this is in the service error invalid phone number")
		err = fmt.Errorf("invalid phone number")
		return err
	}
	fmt.Println("this is the dataaa ", request.Email)
	existingUser, err := s.repo.Login(ctx, request.Email)
	fmt.Println("there may be a user", existingUser)
	if err != nil && err != gorm.ErrRecordNotFound {
		fmt.Println("this is in the service error checking existing user")
		err = fmt.Errorf("failed to check existing user: %w", err)
		return err
	}
	if existingUser.Email != "" {
		fmt.Println("this is in the service user already exists")
		err = fmt.Errorf("user already exists")
		return err
	}
	fmt.Println("this is in the service Register", request.Password)
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(request.Password), bcrypt.DefaultCost)
	if err != nil {
		fmt.Println("this is in the service error hashing password")
		err = fmt.Errorf("failed to hash password: %w", err)
		return err
	}
	request.Password = string(hashedPassword)
	fmt.Println("this is in the service Register", request.Password)
	id, _ := s.repo.Register(ctx, request)
	if err != nil {
		return fmt.Errorf("failed to register user: %w", err)
	}
	return s.repo.CreateWallet(ctx, id)
}

func (s *service) Login(ctx context.Context, request model.UserLoginRequest) error {
	fmt.Println("this is in the service Login", request.Password)
	var err error
	if request.Email == "" || request.Password == "" {
		fmt.Println("this is in the service error value missing")
		err = fmt.Errorf("missing values")
		return err
	}
	storedUser, err := s.repo.Login(ctx, request.Email)
	fmt.Println("thisss is the dataaa ", storedUser)
	if err != nil {
		fmt.Println("this is in the service user not found")
		return fmt.Errorf("user not found: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(storedUser.Password), []byte(request.Password)); err != nil {
		fmt.Println("this is in the service incorrect password")
		return fmt.Errorf("incorrect password: %w", err)
	}

	return nil
}

func (s *service) OtpLogin(ctx context.Context, request model.UserOtp) error {
	fmt.Println("this is in the service Login", request.Otp)
	var err error
	if request.Email == "" || request.Otp == "" {
		fmt.Println("this is in the service error value missing")
		err = fmt.Errorf("missing values")
		return err
	}
	return nil
}

func (s *service) UpdateUser(ctx context.Context, updatedData model.UserRegisterRequest) error {
	var query string
	var args []interface{}

	query = "UPDATE users SET"

	if updatedData.FirstName != "" {
		query += " firstname = ?,"
		args = append(args, updatedData.FirstName)
	}
	if updatedData.LastName != "" {
		query += " lastname = ?,"
		args = append(args, updatedData.LastName)
	}
	if updatedData.Password != "" {
		// Hash the password before updating it
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(updatedData.Password), bcrypt.DefaultCost)
		if err != nil {
			return fmt.Errorf("failed to hash password: %w", err)
		}
		query += " password = ?,"
		args = append(args, string(hashedPassword))
	}
	if updatedData.Phone != "" && isValidPhoneNumber(updatedData.Phone) {
		query += " phone = ?,"
		args = append(args, updatedData.Phone)
	}

	query = strings.TrimSuffix(query, ",")

	query += " WHERE email = ?"
	args = append(args, updatedData.Email)
	fmt.Println("this is the UpdateUser ", query, " kkk ", args)

	return s.repo.UpdateUser(ctx, query, args)
}

func (s *service) Listing(ctx context.Context) ([]model.ProductListingUsers, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		return s.repo.Listing(ctx)
	}
}
func (s *service) ListingSingle(ctx context.Context, id string) ([]model.ProductListDetailed, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		return s.repo.ListingSingle(ctx, id)
	}
}
func (s *service) CategoryListing(ctx context.Context, category string) ([]model.ProductListingUsers, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		return s.repo.CategoryListing(ctx, category)
	}
}
func (s *service) Listcart(ctx context.Context, id string) ([]model.Usercartview, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		return s.repo.Listcart(ctx, id)
	}
}
func (s *service) ListWish(ctx context.Context, id string) ([]model.UserWishview, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		return s.repo.ListWish(ctx, id)
	}
}
func (s *service) LatestListing(ctx context.Context) ([]model.ProductListingUsers, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		return s.repo.LatestListing(ctx)
	}
}
func (s *service) PhighListing(ctx context.Context) ([]model.ProductListingUsers, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		return s.repo.PhighListing(ctx)
	}
}
func (s *service) PlowListing(ctx context.Context) ([]model.ProductListingUsers, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		return s.repo.PlowListing(ctx)
	}
}
func (s *service) InAZListing(ctx context.Context) ([]model.ProductListingUsers, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		return s.repo.InAZListing(ctx)
	}
}

func (s *service) InZAListing(ctx context.Context) ([]model.ProductListingUsers, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		return s.repo.InZAListing(ctx)
	}
}
func isValidEmail(email string) bool {
	// Simple regex pattern for basic email validation
	fmt.Println(" check email validity")
	const emailRegex = `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	re := regexp.MustCompile(emailRegex)
	return re.MatchString(email)
}
func isValidPhoneNumber(phone string) bool {
	// Simple regex pattern for basic phone number validation
	fmt.Println(" check pfone validity")
	const phoneRegex = `^\+?[1-9]\d{1,14}$` // E.164 international phone number format
	re := regexp.MustCompile(phoneRegex)
	return re.MatchString(phone)
}
func (s *service) BrandListing(ctx context.Context, category string) ([]model.ProductListingUsers, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		return s.repo.BrandListing(ctx, category)
	}
}

func (s *service) BestSellingListingProductCategory(ctx context.Context, category string) ([]model.ProductListingUsers, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		return s.repo.BestSellingListingProductCategory(ctx, category)
	}
}

func (s *service) BestSellingListingProductBrand(ctx context.Context, category string) ([]model.ProductListingUsers, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		return s.repo.BestSellingListingProductBrand(ctx, category)
	}
}
func (s *service) BestSellingListingProduct(ctx context.Context) ([]model.ProductListingUsers, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		return s.repo.BestSellingListingProduct(ctx)
	}
}

func (s *service) BestSellingListingCategory(ctx context.Context) ([]string, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		return s.repo.BestSellingListingCategory(ctx)
	}
}
func (s *service) BestSellingListingBrand(ctx context.Context) ([]string, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		return s.repo.BestSellingListingBrand(ctx)
	}
}
