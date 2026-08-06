import { useState } from "react";
import { Link, useNavigate } from "react-router-dom";
import { useTranslation } from "react-i18next";
import { Form, Input } from "antd";
import { successToast, errorToast } from "../../utils/utils";
import LoadableButton from "../../components/button/LoadableButton";
import OtpVerify from "./components/OtpVerify";
import {
  forgetPasswordSendOtp,
  forgetPasswordVerifyOtp,
  forgotPassword,
} from "../../api/request/auth";
import MESSAGES from "../../utils/messages";
import Logo from "../../components/common/Logo";

const ForgotPassword = () => {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const [step, setStep] = useState(1);
  const [email, setEmail] = useState("");
  const [otp, setOtp] = useState("");
  const [isLoading, setIsLoading] = useState(false);

  const handleEmailSubmit = async (values) => {
    setIsLoading(true);
    try {
      const body = { email: values?.email };
      const { data, status } = await forgetPasswordSendOtp(body);
      if (status >= 200 && status < 300) {
        successToast(data?.message || MESSAGES.OTP_SENT_SUCCESS);
        setEmail(values?.email);
        setStep(2);
      }
    } catch (error) {
      errorToast(error?.response?.data?.message);
    } finally {
      setIsLoading(false);
    }
  };

  const handleOtpSubmit = async (values) => {
    setIsLoading(true);
    try {
      const body = { email, code: values?.otp };
      const { data, status } = await forgetPasswordVerifyOtp(body);
      if (status >= 200 && status < 300) {
        successToast(data?.message || MESSAGES.OTP_VERIFICATION_SUCCESS);
        setOtp(values?.otp);
        setStep(3);
      }
    } catch (error) {
      errorToast(error?.response?.data?.message);
    } finally {
      setIsLoading(false);
    }
  };

  const handlePasswordSubmit = async (values) => {
    setIsLoading(true);
    try {
      const body = {
        email,
        code: otp,
        password: values?.newPassword,
      };
      const { data, status } = await forgotPassword(body);
      if (status >= 200 && status < 300) {
        successToast(data?.message || MESSAGES.PASSWORD_RESET_SUCCESS);
        navigate("/login");
      }
    } catch (error) {
      errorToast(error?.response?.data?.message);
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div className="auth-card">
      <div className="flex items-center justify-center mb-5">
        <Logo className="h-16" />
      </div>

      {step === 1 && (
        <Form layout="vertical" onFinish={handleEmailSubmit}>
          <Form.Item
            label={t("username")}
            name="email"
            rules={[{ required: true, message: "Please enter your username" }]}
          >
            <Input placeholder={t("enterUsername")} />
          </Form.Item>

          <LoadableButton
            type="submit"
            lable={t("sendOtp")}
            className="w-full"
            isLoading={isLoading}
            loadingLable={t("sendingOtp")}
          />
        </Form>
      )}

      {step === 2 && (
        <>
          <div className="text-center mb-5">
            <p className="text-2xl font-semibold mb-2">{t("verifyOtp")}</p>
          </div>

          <OtpVerify
            email={email}
            handleSubmit={handleOtpSubmit}
            isLoading={isLoading}
          />
        </>
      )}

      {step === 3 && (
        <>
          <div className="text-center mb-5">
            <p className="text-2xl font-semibold mb-2">{t("resetPassword")}</p>
          </div>

          <Form layout="vertical" onFinish={handlePasswordSubmit}>
            <Form.Item
              label={t("newPassword")}
              name="newPassword"
              rules={[
                { required: true, message: "Please enter new password" },
                {
                  pattern:
                    /^(?=.*[A-Z])(?=.*\d)(?=.*[!@#$%^&*])[A-Za-z\d!@#$%^&*]{8,}$/,
                  message: "Please enter strong password",
                },
              ]}
            >
              <Input.Password placeholder={t("enterNewPassword")} />
            </Form.Item>

            <Form.Item
              label={t("confirmPassword")}
              name="confirmPassword"
              dependencies={["newPassword"]}
              rules={[
                { required: true, message: "Please confirm your password" },
                ({ getFieldValue }) => ({
                  validator(_, value) {
                    if (!value || getFieldValue("newPassword") === value) {
                      return Promise.resolve();
                    }
                    return Promise.reject(new Error("Passwords do not match"));
                  },
                }),
              ]}
            >
              <Input.Password placeholder={t("enterConfirmPassword")} />
            </Form.Item>

            <LoadableButton
              type="submit"
              lable={t("resetPassword")}
              className="w-full"
              isLoading={isLoading}
              loadingLable={t("resettingPassword")}
            />
          </Form>
        </>
      )}

      <p className="text-center mt-5">
        <Link to="/login" className="!text-primary">
          {t("backToLogin")}
        </Link>
      </p>
    </div>
  );
};

export default ForgotPassword;
