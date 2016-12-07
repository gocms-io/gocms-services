package auth

import (
	"net/http"
	"github.com/gin-gonic/gin"
	"github.com/menklab/goCMS/utility/errors"
	"log"
)

type ResetPasswordRequest struct {
	Email string `json:"email" binding:"required"`
}

type ResetPassword struct {
	Email     string `json:"email" binding:"required"`
	Password  string `json:"password" binding:"required"`
	ResetCode string `json:"resetCode" binding:"required"`
}



/**
* @api {post} /reset-password Reset Password (Request)
* @apiName ResetPassword
* @apiGroup Authentication
*
* @apiUse AuthHeader
*
* @apiParam {string} email
*
* @apiSuccessExample Success-Response:
*     HTTP/1.1 200 OK
*
* @apiErrorExample Error-Response:
*     HTTP/1.1 403 Unauthorized
*/
func (ac *AuthController) resetPassword(c *gin.Context) {

	// get email for reset
	var resetRequest ResetPasswordRequest
	err := c.BindJSON(&resetRequest) // update any changes from request
	if err != nil {
		errors.Response(c, http.StatusBadRequest, "Missing Fields", err)
		return
	}

	// send password reset link
	err = ac.ServicesGroup.AuthService.SendPasswordResetCode(resetRequest.Email)
	if err != nil {
		log.Printf("Error sending reset email: %s", err.Error())
		errors.Response(c, http.StatusInternalServerError, errors.ApiError_Server, err)
		return
	}

	// respond as everything after this doesn't matter to the requester
	c.String(http.StatusOK, "Email will be sent to the account provided.")
}

/**
* @api {put} /reset-password Reset Password (Verify/Set)
* @apiName SetResetPassword
* @apiGroup Authentication
*
* @apiUse AuthHeader
*
* @apiParam {string} email
* @apiParam {string} password
* @apiParam {string} resetCode
*
* @apiSuccessExample Success-Response:
*     HTTP/1.1 200 OK
*
* @apiErrorExample Error-Response:
*     HTTP/1.1 403 Unauthorized
*/
func (ac *AuthController) setPassword(c *gin.Context) {
	// get password and code for reset
	var resetPassword ResetPassword
	err := c.BindJSON(&resetPassword) // update any changes from request
	if err != nil {
		errors.Response(c, http.StatusBadRequest, err.Error(), err)
		return
	}

	// get user
	user, err := ac.ServicesGroup.UserService.GetByEmail(resetPassword.Email)
	if err != nil {
		errors.Response(c, http.StatusBadRequest, "Couldn't reset password.", err)
		return
	}

	// verify code
	if ok := ac.ServicesGroup.AuthService.VerifyPasswordResetCode(user.Id, resetPassword.ResetCode); !ok {
		errors.ResponseWithSoftRedirect(c, http.StatusUnauthorized, "Error resetting password.", REDIRECT_LOGIN)
		return
	}

	// reset password
	err = ac.ServicesGroup.UserService.UpdatePassword(user.Id, resetPassword.Password)
	if err != nil {
		errors.Response(c, http.StatusBadRequest, "Couldn't reset password.", err)
		return
	}

	c.Status(http.StatusOK)
}
