package ucases

import (
	"context"
	"errors"
	"testing"
	"time"

	"go.uber.org/mock/gomock"

	"github.com/lifenetwork-ai/iam-service/constants"
	cachingtypes "github.com/lifenetwork-ai/iam-service/infrastructures/caching/types"
	otpqueue "github.com/lifenetwork-ai/iam-service/infrastructures/otp_queue/types"
	mock_sms "github.com/lifenetwork-ai/iam-service/mocks/domain/ucases/services"
	mock_cache "github.com/lifenetwork-ai/iam-service/mocks/infrastructures/caching/types"
	mock_queue "github.com/lifenetwork-ai/iam-service/mocks/infrastructures/otp_queue/types"
	"github.com/stretchr/testify/assert"
)

func newCourier(ctrl *gomock.Controller) (*courierUseCase, *mock_queue.MockOTPQueueRepository, *mock_sms.MockSMSProvider, *mock_cache.MockCacheRepository) {
	q := mock_queue.NewMockOTPQueueRepository(ctrl)
	sms := mock_sms.NewMockSMSProvider(ctrl)
	cache := mock_cache.NewMockCacheRepository(ctrl)
	uc := NewCourierUseCase(q, sms, cache).(*courierUseCase)
	return uc, q, sms, cache
}

func TestCourier_ChooseChannel_Success_And_NotSupported(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	uc, _, _, cache := newCourier(ctrl)

	ctx := context.Background()
	// Supported: genetica, +84 => includes zalo
	cache.EXPECT().SaveItem(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
	err := uc.ChooseChannel(ctx, constants.TenantGenetica, "+84123456789", constants.ChannelZalo)
	assert.Nil(t, err)

	// Not supported channel
	err = uc.ChooseChannel(ctx, constants.TenantLifeAI, "+12025550123", "telegram")
	assert.NotNil(t, err)
}

func TestCourier_GetChannel_Hit_And_Miss(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	uc, _, _, cache := newCourier(ctrl)

	ctx := context.Background()
	var stored string = constants.ChannelSMS
	cache.EXPECT().RetrieveItem(gomock.Any(), gomock.Any()).DoAndReturn(
		func(key *cachingtypes.Keyer, out any) error {
			p := out.(*string)
			*p = stored
			return nil
		},
	)
	resp, err := uc.GetChannel(ctx, "tenant", "+12025550123")
	assert.Nil(t, err)
	assert.Equal(t, constants.ChannelSMS, resp.Channel)

	// miss path
	cacheMissCtrl := gomock.NewController(t)
	defer cacheMissCtrl.Finish()
	uc2, _, _, cache2 := newCourier(cacheMissCtrl)
	cache2.EXPECT().RetrieveItem(gomock.Any(), gomock.Any()).Return(cachingtypes.ErrCacheMiss)
	resp, err = uc2.GetChannel(ctx, "tenant", "+12025550123")
	assert.Nil(t, err)
	assert.Equal(t, "mock", resp.Channel)
}

func TestCourier_GetChannel_InternalError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	uc, _, _, cache := newCourier(ctrl)

	x := errors.New("boom")
	cache.EXPECT().RetrieveItem(gomock.Any(), gomock.Any()).Return(x)
	resp, derr := uc.GetChannel(context.Background(), "tenant", "+12025550123")
	assert.NotNil(t, derr)
	assert.Equal(t, "", resp.Channel)
}

func TestCourier_ReceiveOTP_Validations_And_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	uc, queue, _, _ := newCourier(ctrl)
	ctx := context.Background()

	// invalid tenant (no [TENANT])
	err := uc.ReceiveOTP(ctx, "+12025550123", "Your code 123456")
	assert.NotNil(t, err)

	// invalid otp (no 6-digit)
	err = uc.ReceiveOTP(ctx, "+12025550123", "[LIFE AI] no code here")
	assert.NotNil(t, err)

	// invalid tenant value
	err = uc.ReceiveOTP(ctx, "+12025550123", "[bad] code 123456")
	assert.NotNil(t, err)

	// success uses bracketed tenant
	queue.EXPECT().Enqueue(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
	err = uc.ReceiveOTP(ctx, "+12025550123", "[life_ai] Your verification code is 123456")
	assert.Nil(t, err)
}

func TestCourier_DeliverOTP_Success_And_Retry(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	uc, queue, sms, cache := newCourier(ctrl)
	ctx := context.Background()

	// choose channel in cache
	cache.EXPECT().RetrieveItem(gomock.Any(), gomock.Any()).DoAndReturn(
		func(key *cachingtypes.Keyer, out any) error {
			p := out.(*string)
			*p = constants.ChannelSMS
			return nil
		})

	// success path
	queue.EXPECT().Get(gomock.Any(), "tenant", "+12025550123").Return(&otpqueue.OTPQueueItem{Message: "123456"}, nil)
	sms.EXPECT().SendOTP(gomock.Any(), "tenant", "+12025550123", constants.ChannelSMS, "123456", gomock.Any()).Return(nil)
	queue.EXPECT().Delete(gomock.Any(), "tenant", "+12025550123").Return(nil)
	err := uc.DeliverOTP(ctx, "tenant", "+12025550123")
	assert.Nil(t, err)

	// retry path: send fails -> enqueue retry
	cache.EXPECT().RetrieveItem(gomock.Any(), gomock.Any()).DoAndReturn(
		func(key *cachingtypes.Keyer, out any) error {
			p := out.(*string)
			*p = constants.ChannelWhatsApp
			return nil
		})
	queue.EXPECT().Get(gomock.Any(), "tenant", "+12025550124").Return(&otpqueue.OTPQueueItem{Message: "999999"}, nil)
	sms.EXPECT().SendOTP(gomock.Any(), "tenant", "+12025550124", constants.ChannelWhatsApp, "999999", gomock.Any()).Return(errors.New("fail"))
	queue.EXPECT().EnqueueRetry(gomock.Any(), gomock.AssignableToTypeOf(otpqueue.RetryTask{})).Return(nil)
	queue.EXPECT().Delete(gomock.Any(), "tenant", "+12025550124").Return(nil)
	err = uc.DeliverOTP(ctx, "tenant", "+12025550124")
	assert.NotNil(t, err)
}

func TestCourier_DeliverOTP_QueueGetError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	uc, queue, _, cache := newCourier(ctrl)

	cache.EXPECT().RetrieveItem(gomock.Any(), gomock.Any()).DoAndReturn(
		func(key *cachingtypes.Keyer, out any) error {
			p := out.(*string)
			*p = constants.ChannelSMS
			return nil
		})
	queue.EXPECT().Get(gomock.Any(), "tenant", "+12025550000").Return(nil, errors.New("no item"))
	err := uc.DeliverOTP(context.Background(), "tenant", "+12025550000")
	assert.NotNil(t, err)
}

func TestCourier_RetryFailedOTPs_Behavior(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	uc, queue, sms, _ := newCourier(ctrl)
	ctx := context.Background()
	now := time.Now()

	tasks := []otpqueue.RetryTask{
		{Receiver: "+12025550001", Message: "111111", Channel: constants.ChannelSMS, TenantName: "tenant", RetryCount: 0},
		{Receiver: "+12025550002", Message: "222222", Channel: constants.ChannelWhatsApp, TenantName: "tenant", RetryCount: constants.MaxOTPRetryCount},
	}
	queue.EXPECT().GetDueRetryTasks(gomock.Any(), now).Return(tasks, nil)
	// first succeed, second exceed max -> both delete
	sms.EXPECT().SendOTP(gomock.Any(), "tenant", "+12025550001", constants.ChannelSMS, "111111", gomock.Any()).Return(nil)
	sms.EXPECT().SendOTP(gomock.Any(), "tenant", "+12025550002", constants.ChannelWhatsApp, "222222", gomock.Any()).Return(errors.New("fail"))
	queue.EXPECT().DeleteRetryTask(gomock.Any(), tasks[0]).Return(nil)
	queue.EXPECT().DeleteRetryTask(gomock.Any(), tasks[1]).Return(nil)
	queue.EXPECT().Delete(gomock.Any(), "tenant", "+12025550001").Return(nil).AnyTimes()
	queue.EXPECT().Delete(gomock.Any(), "tenant", "+12025550002").Return(nil).AnyTimes()

	count, derr := uc.RetryFailedOTPs(ctx, now)
	assert.Nil(t, derr)
	assert.Equal(t, 2, count)
}
