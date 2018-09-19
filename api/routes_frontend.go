// CreatePinPayment is used to create a signed message for a pin payment
func (api *API) createPinPayment(c *gin.Context) {
	contentHash := c.Param("hash")
	if _, err := gocid.Decode(contentHash); err != nil {
		FailOnError(c, err)
		return
	}
	username := GetAuthenticatedUserFromContext(c)
	holdTime, exists := c.GetPostForm("hold_time")
	if !exists {
		FailNoExistPostForm(c, "hold_time")
		return
	}
	method, exists := c.GetPostForm("payment_method")
	if !exists {
		FailNoExistPostForm(c, "payment_method")
		return
	}
	methodUint, err := strconv.ParseUint(method, 10, 8)
	if err != nil {
		FailOnError(c, err)
		return
	}
	if methodUint > 1 {
		FailOnError(c, errors.New("payment_method must be 1 or 0"))
		return
	}
	holdTimeInt, err := strconv.ParseInt(holdTime, 10, 64)
	if err != nil {
		FailOnError(c, err)
		return
	}

	manager, err := rtfs.Initialize("", "")
	if err != nil {
		api.LogError(err, IPFSConnectionError)
		FailOnError(c, err)
		return
	}
	totalCost, err := utils.CalculatePinCost(contentHash, holdTimeInt, manager.Shell)
	if err != nil {
		api.LogError(err, PinCostCalculationError)
		FailOnError(c, err)
		return
	}

	keyFile := api.TConfig.Ethereum.Account.KeyFile
	keyPass := api.TConfig.Ethereum.Account.KeyPass
	ps, err := signer.GeneratePaymentSigner(keyFile, keyPass)
	if err != nil {
		api.LogError(err, PaymentSignerGenerationError)
		FailOnError(c, err)
		return
	}
	um := models.NewUserManager(api.DBM.DB)
	ethAddress, err := um.FindEthAddressByUserName(username)
	if err != nil {
		api.LogError(err, EthAddressSearchError)
		FailOnError(c, err)
		return
	}
	ppm := models.NewPaymentManager(api.DBM.DB)
	var num *big.Int
	num, err = ppm.RetrieveLatestPaymentNumberForUser(username)
	if err != nil && err != gorm.ErrRecordNotFound {
		api.LogError(err, PaymentSearchError)
		FailOnError(c, err)
		return
	}
	if num == nil {
		num = big.NewInt(0)
	}
	num = new(big.Int).Add(num, big.NewInt(1))
	costBig := utils.FloatToBigInt(totalCost)
	// for testing purpose
	addressTyped := common.HexToAddress(ethAddress)

	sm, err := ps.GenerateSignedPaymentMessagePrefixed(addressTyped, uint8(methodUint), num, costBig)
	if err != nil {
		api.LogError(err, PaymentMessageSignError)
		FailOnError(c, err)
		return
	}

	if _, err = ppm.NewPayment(uint8(methodUint), sm.PaymentNumber, sm.ChargeAmount, ethAddress, contentHash, username, "pin", "public", holdTimeInt); err != nil {
		api.LogError(err, PaymentCreationError)
		FailOnError(c, err)
		return
	}

	api.Logger.WithFields(log.Fields{
		"service":        "api",
		"user":           username,
		"payment_number": sm.PaymentNumber.String(),
	}).Info("pin payment request generated")

	Respond(c, http.StatusOK, gin.H{"response": gin.H{
		"h":                    sm.H,
		"v":                    sm.V,
		"r":                    sm.R,
		"s":                    sm.S,
		"eth_address":          sm.Address,
		"charge_amount_in_wei": sm.ChargeAmount,
		"payment_method":       sm.PaymentMethod,
		"payment_number":       sm.PaymentNumber}})
}

// CreateFilePayment is used to create a signed file payment message
func (api *API) createFilePayment(c *gin.Context) {
	networkName, exists := c.GetPostForm("network_name")
	if !exists {
		FailNoExistPostForm(c, "network_name")
		return
	}
	holdTimeInMonths, exists := c.GetPostForm("hold_time")
	if !exists {
		FailNoExistPostForm(c, "hold_time")
		return
	}

	method, exists := c.GetPostForm("payment_method")
	if !exists {
		FailNoExistPostForm(c, "payment_method")
		return
	}
	fileHandler, err := c.FormFile("file")
	if err != nil {
		FailOnError(c, err)
		return
	}
	if err := api.FileSizeCheck(fileHandler.Size); err != nil {
		FailOnError(c, err)
		return
	}
	methodUint, err := strconv.ParseUint(method, 10, 8)
	if err != nil {
		FailOnError(c, err)
		return
	}
	if methodUint > 1 {
		FailOnError(c, errors.New("payment_method must be 1 or 0"))
		return
	}

	accessKey := api.TConfig.MINIO.AccessKey
	secretKey := api.TConfig.MINIO.SecretKey
	endpoint := fmt.Sprintf("%s:%s", api.TConfig.MINIO.Connection.IP, api.TConfig.MINIO.Connection.Port)
	keyFile := api.TConfig.Ethereum.Account.KeyFile
	keyPass := api.TConfig.Ethereum.Account.KeyPass
	ps, err := signer.GeneratePaymentSigner(keyFile, keyPass)
	if err != nil {
		api.LogError(err, PaymentSignerGenerationError)
		FailOnError(c, err)
		return
	}
	miniManager, err := mini.NewMinioManager(endpoint, accessKey, secretKey, false)
	if err != nil {
		api.LogError(err, MinioConnectionError)
		FailOnError(c, err)
		return
	}

	fmt.Println("opening file")
	openFile, err := fileHandler.Open()
	if err != nil {
		api.LogError(err, FileOpenError)
		FailOnError(c, err)
		return
	}
	fmt.Println("file opened")
	username := GetAuthenticatedUserFromContext(c)

	holdTimeInMonthsInt, err := strconv.ParseInt(holdTimeInMonths, 10, 64)
	if err != nil {
		FailOnError(c, err)
		return
	}
	cost := utils.CalculateFileCost(holdTimeInMonthsInt, fileHandler.Size)
	costBig := utils.FloatToBigInt(cost)
	randUtils := utils.GenerateRandomUtils()
	randString := randUtils.GenerateString(32, utils.LetterBytes)
	objectName := fmt.Sprintf("%s%s", username, randString)
	fmt.Println("storing file in minio")
	if _, err = miniManager.PutObject(FilesUploadBucket, objectName, openFile, fileHandler.Size, minio.PutObjectOptions{}); err != nil {
		api.LogError(err, MinioPutError)
		FailOnError(c, err)
		return
	}
	fmt.Println("file stored in minio")

	pm := models.NewPaymentManager(api.DBM.DB)
	um := models.NewUserManager(api.DBM.DB)
	ethAddress, err := um.FindEthAddressByUserName(username)
	if err != nil {
		api.LogError(err, EthAddressSearchError)
		FailOnError(c, err)
		return
	}
	var num *big.Int
	num, err = pm.RetrieveLatestPaymentNumberForUser(username)
	if err != nil && err != gorm.ErrRecordNotFound {
		api.LogError(err, PaymentSearchError)
		FailOnError(c, err)
		return
	}
	if num == nil {
		num = big.NewInt(0)

	} else if num.Cmp(big.NewInt(0)) == 1 {
		// this means that the latest payment number is greater than 0
		// indicating a payment has already been made, in which case
		// we will increment the value by 1
		num = new(big.Int).Add(num, big.NewInt(1))
	}
	addressTyped := common.HexToAddress(ethAddress)
	sm, err := ps.GenerateSignedPaymentMessagePrefixed(addressTyped, uint8(methodUint), num, costBig)
	if err != nil {
		api.LogError(err, PaymentMessageSignError)
		FailOnError(c, err)
		return
	}
	if _, err = pm.NewPayment(uint8(methodUint), sm.PaymentNumber, sm.ChargeAmount, ethAddress, objectName, username, "file", networkName, holdTimeInMonthsInt); err != nil {
		api.LogError(err, PaymentCreationError)
		FailOnError(c, err)
		return
	}

	api.Logger.WithFields(log.Fields{
		"service":        "api",
		"user":           username,
		"payment_number": sm.PaymentNumber.String(),
	}).Info("file payment request generated")

	Respond(c, http.StatusOK, gin.H{"response": gin.H{
		"h":                    sm.H,
		"v":                    sm.V,
		"r":                    sm.R,
		"s":                    sm.S,
		"eth_address":          sm.Address,
		"charge_amount_in_wei": sm.ChargeAmount,
		"payment_method":       sm.PaymentMethod,
		"payment_number":       sm.PaymentNumber}})
}

// SubmitPinPaymentConfirmation is used to submit a pin payment confirmationrequest to the backend.
// A successful payment will result in the content being injected into temporal
func (api *API) submitPinPaymentConfirmation(c *gin.Context) {
	username := GetAuthenticatedUserFromContext(c)
	paymentNumber, exists := c.GetPostForm("payment_number")
	if !exists {
		FailNoExistPostForm(c, "payment_number")
		return
	}
	txHash, exists := c.GetPostForm("tx_hash")
	if !exists {
		FailNoExistPostForm(c, "tx_hash")
		return
	}
	ppm := models.NewPaymentManager(api.DBM.DB)
	um := models.NewUserManager(api.DBM.DB)
	ethAddress, err := um.FindEthAddressByUserName(username)
	if err != nil {
		api.LogError(err, EthAddressSearchError)
		FailOnError(c, err)
		return
	}
	pp, err := ppm.FindPaymentByNumberAndAddress(paymentNumber, ethAddress)
	if err != nil {
		api.LogError(err, PaymentSearchError)
		FailOnError(c, err)
		return
	}
	mqURL := api.TConfig.RabbitMQ.URL

	ppc := queue.PinPaymentConfirmation{
		TxHash:        txHash,
		EthAddress:    ethAddress,
		PaymentNumber: paymentNumber,
		ContentHash:   pp.ObjectName,
	}
	qm, err := queue.Initialize(queue.PinPaymentConfirmationQueue, mqURL, true, false)
	if err != nil {
		api.LogError(err, QueueInitializationError)
		FailOnError(c, err)
		return
	}
	fmt.Println("publishing message")
	if err = qm.PublishMessage(ppc); err != nil {
		api.LogError(err, QueuePublishError)
		FailOnError(c, err)
		return
	}
	api.Logger.WithFields(log.Fields{
		"service":        "api",
		"user":           username,
		"payment_number": paymentNumber,
	}).Info("pin payment confirmation being processed")

	Respond(c, http.StatusOK, gin.H{"response": pp})
}

