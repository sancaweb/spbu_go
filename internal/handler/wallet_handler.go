package handler

import (
	"net/http"
	"strconv"
	"strings"

	"spbu_go/internal/entity"
	"spbu_go/internal/service"

	"github.com/gin-gonic/gin"
)

type WalletHandler struct {
	walletService service.WalletService
}

func NewWalletHandler(walletService service.WalletService) *WalletHandler {
	return &WalletHandler{walletService}
}

func (h *WalletHandler) Index(c *gin.Context) {
	user, _ := c.Get("user")
	favicon, _ := c.Get("favicon")
	wallets, _ := h.walletService.GetAll()

	c.HTML(http.StatusOK, "master/keuangan/wallet/index.html", gin.H{
		"Title":      "Wallet / Dompet",
		"ActiveMenu": "keuangan_wallet",
		"User":       user,
		"Favicon":    favicon,
		"Wallets":    wallets,
	})
}

func (h *WalletHandler) Create(c *gin.Context) {
	userVal, _ := c.Get("user")

	walletName := strings.TrimSpace(c.PostForm("wallet_name"))
	description := strings.TrimSpace(c.PostForm("description"))
	isDefault := c.PostForm("is_default") == "true"
	saldoStr := c.PostForm("saldo")
	saldo, _ := strconv.ParseInt(strings.ReplaceAll(saldoStr, ".", ""), 10, 64)

	if walletName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"status": false, "message": "Nama wallet tidak boleh kosong"})
		return
	}

	wallet := &entity.Wallet{
		WalletName:  walletName,
		IsDefault:   isDefault,
		Description: description,
		Saldo:       saldo,
	}
	if u, ok := userVal.(entity.User); ok {
		wallet.UpdatedBy = &u.ID
	}

	if err := h.walletService.Create(wallet); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": false, "message": "Gagal menyimpan wallet: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": true, "message": "Wallet berhasil ditambahkan"})
}

func (h *WalletHandler) Update(c *gin.Context) {
	userVal, _ := c.Get("user")
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": false, "message": "ID tidak valid"})
		return
	}

	walletName := strings.TrimSpace(c.PostForm("wallet_name"))
	description := strings.TrimSpace(c.PostForm("description"))
	isDefault := c.PostForm("is_default") == "true"
	saldoStr := c.PostForm("saldo")
	saldo, _ := strconv.ParseInt(strings.ReplaceAll(saldoStr, ".", ""), 10, 64)

	if walletName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"status": false, "message": "Nama wallet tidak boleh kosong"})
		return
	}

	wallet := &entity.Wallet{
		WalletName:  walletName,
		IsDefault:   isDefault,
		Description: description,
		Saldo:       saldo,
	}
	if u, ok := userVal.(entity.User); ok {
		wallet.UpdatedBy = &u.ID
	}

	if err := h.walletService.Update(uint(id), wallet); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": false, "message": "Gagal memperbarui wallet: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": true, "message": "Wallet berhasil diperbarui"})
}

func (h *WalletHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": false, "message": "ID tidak valid"})
		return
	}
	if err := h.walletService.Delete(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": false, "message": "Gagal menghapus wallet: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": true, "message": "Wallet berhasil dihapus"})
}
