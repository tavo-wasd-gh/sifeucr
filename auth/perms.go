package auth

const (
	// Usuario
	CreateUsuario = 1 << iota
	DeleteUsuario
	ReadUsuarioOther
	WriteUsuarioOther
	ReadUsuarioOwn
	WriteUsuarioOwn
	// Cuentas
	CreateCuenta
	DeleteCuenta
	ReadCuentaOther
	WriteCuentaOther
	ReadCuentaOwn
	WriteCuentaOwn
	// Presupuesto
	CreatePresupuesto
	DeletePresupuesto
	ReadPresupuestoOther
	WritePresupuestoOther
	ReadPresupuestoOwn
	WritePresupuestoOwn
	// Servicios
	CreateServicios
	DeleteServicios
	ReadServiciosOther
	WriteServiciosOther
	ReadServiciosOwn
	WriteServiciosOwn
	// Suministros
	CreateSuministros
	DeleteSuministros
	ReadSuministrosOther
	WriteSuministrosOther
	ReadSuministrosOwn
	WriteSuministrosOwn
	// Bienes
	CreateBienes
	DeleteBienes
	ReadBienesOther
	WriteBienesOther
	ReadBienesOwn
	WriteBienesOwn
	// Ajustes
	CreateAjustes
	DeleteAjustes
	ReadAjustesOther
	WriteAjustesOther
	ReadAjustesOwn
	WriteAjustesOwn
	// Donaciones
	CreateDonaciones
	DeleteDonaciones
	ReadDonacionesOther
	WriteDonacionesOther
	ReadDonacionesOwn
	WriteDonacionesOwn
	// Defined
	Regular =
		ReadUsuarioOwn |
		ReadCuentaOwn |
		ReadPresupuestoOwn |
		CreateServicios |
		ReadServiciosOwn |
		WriteServiciosOwn |
		CreateSuministros |
		ReadSuministrosOwn |
		WriteSuministrosOwn |
		CreateBienes|
		ReadBienesOwn |
		WriteBienesOwn |
		ReadAjustesOwn |
		CreateDonaciones |
		ReadDonacionesOwn |
		WriteDonacionesOwn
	SF =
		CreateUsuario |
		DeleteUsuario |
		ReadUsuarioOther |
		WriteUsuarioOther |
		ReadUsuarioOwn |
		WriteUsuarioOwn |
		CreateCuenta |
		DeleteCuenta |
		ReadCuentaOther |
		WriteCuentaOther |
		ReadCuentaOwn |
		WriteCuentaOwn |
		CreatePresupuesto |
		DeletePresupuesto |
		ReadPresupuestoOther |
		WritePresupuestoOther |
		ReadPresupuestoOwn |
		WritePresupuestoOwn |
		CreateServicios |
		DeleteServicios |
		ReadServiciosOther |
		WriteServiciosOther |
		ReadServiciosOwn |
		WriteServiciosOwn |
		CreateSuministros |
		DeleteSuministros |
		ReadSuministrosOther |
		WriteSuministrosOther |
		ReadSuministrosOwn |
		WriteSuministrosOwn |
		CreateBienes |
		DeleteBienes |
		ReadBienesOther |
		WriteBienesOther |
		ReadBienesOwn |
		WriteBienesOwn |
		CreateAjustes |
		DeleteAjustes |
		ReadAjustesOther |
		WriteAjustesOther |
		ReadAjustesOwn |
		WriteAjustesOwn |
		ReadDonacionesOther |
		ReadDonacionesOwn
	COES =
		ReadUsuarioOwn |
		ReadCuentaOwn |
		ReadPresupuestoOwn |
		ReadPresupuestoOther |
		CreateServicios |
		ReadServiciosOther |
		WriteServiciosOther |
		ReadServiciosOwn |
		WriteServiciosOwn |
		CreateSuministros |
		ReadSuministrosOther |
		WriteSuministrosOther |
		ReadSuministrosOwn |
		WriteSuministrosOwn |
		CreateBienes|
		ReadBienesOther |
		WriteBienesOther |
		ReadBienesOwn |
		WriteBienesOwn |
		ReadAjustesOther |
		ReadAjustesOwn |
		CreateDonaciones |
		DeleteDonaciones |
		ReadDonacionesOther |
		WriteDonacionesOther |
		ReadDonacionesOwn |
		WriteDonacionesOwn
)
