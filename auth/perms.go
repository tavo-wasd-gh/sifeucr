package auth

const (
	// Users
	CreateUser = 1 << iota
	DeleteUser
	ReadUserOther
	WriteUserOther
	ReadUserOwn
	WriteUserOwn
	// Accounts
	CreateAccount
	DeleteAccount
	ReadAccountOther
	WriteAccountOther
	ReadAccountOwn
	WriteAccountOwn
	// Permissions
	CreatePermission
	DeletePermission
	ReadPermissionOther
	WritePermissionOther
	ReadPermissionOwn
	WritePermissionOwn
	// Budgets
	CreateBudget
	DeleteBudget
	ReadBudgetOther
	WriteBudgetOther
	ReadBudgetOwn
	WriteBudgetOwn
	// Requests
	CreateRequest
	DeleteRequest
	ReadRequestOther
	WriteRequestOther
	ReadRequestOwn
	WriteRequestOwn
	// Adjustments
	CreateAdjustment
	DeleteAdjustment
	ReadAdjustmentOther
	WriteAdjustmentOther
	ReadAdjustmentOwn
	WriteAdjustmentOwn
	// Donations
	CreateDonation
	DeleteDonation
	ReadDonationOther
	WriteDonationOther
	ReadDonationOwn
	WriteDonationOwn
	// Regular
	Regular = CreateUser |
		ReadUserOwn |
		ReadAccountOwn |
		CreatePermission |
		ReadPermissionOwn |
		ReadBudgetOwn |
		CreateRequest |
		ReadRequestOwn |
		ReadAdjustmentOwn |
		CreateDonation |
		ReadDonationOwn
	// Advanced
	Advanced = CreateUser |
		DeleteUser |
		ReadUserOther |
		WriteUserOther |
		ReadUserOwn |
		WriteUserOwn |
		CreateAccount |
		DeleteAccount |
		ReadAccountOther |
		WriteAccountOther |
		ReadAccountOwn |
		WriteAccountOwn |
		CreatePermission |
		DeletePermission |
		ReadPermissionOther |
		WritePermissionOther |
		ReadPermissionOwn |
		WritePermissionOwn |
		CreateBudget |
		DeleteBudget |
		ReadBudgetOther |
		WriteBudgetOther |
		ReadBudgetOwn |
		WriteBudgetOwn |
		CreateRequest |
		DeleteRequest |
		ReadRequestOther |
		WriteRequestOther |
		ReadRequestOwn |
		WriteRequestOwn |
		CreateAdjustment |
		DeleteAdjustment |
		ReadAdjustmentOther |
		WriteAdjustmentOther |
		ReadAdjustmentOwn |
		WriteAdjustmentOwn |
		CreateDonation |
		DeleteDonation |
		ReadDonationOther |
		WriteDonationOther |
		ReadDonationOwn |
		WriteDonationOwn
)
