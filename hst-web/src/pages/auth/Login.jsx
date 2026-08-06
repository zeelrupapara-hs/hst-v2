import { useState } from "react";
import { Link, useNavigate } from "react-router-dom";
import { useTranslation } from "react-i18next";
import { Divider, Form, Input } from "antd";
import useAuthStore from "../../store/useAuthStore";
import LoadableButton from "../../components/button/LoadableButton";
import { login } from "../../api/request/auth";
import { successToast, errorToast, setLocalItem } from "../../utils/utils";
import MESSAGES from "../../utils/messages";
import Logo from "../../components/common/Logo";

const Login = () => {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const { setIsLogin, setUser } = useAuthStore((state) => state);
  const [isLoading, setIsLoading] = useState(false);

  const handleSubmit = async (values) => {
    setIsLoading(true);
    try {
      const { username, password } = values;

      const token = btoa(`${username}:${password}`);
      const headers = {
        headers: {
          Authorization: `Basic ${token}`,
          "Content-Type": "application/json",
        },
      };

      const { data } = await login(headers);

      if (data?.success) {
        const { access_token, refresh_token } = data?.data;
        successToast(data?.message || MESSAGES.LOGIN_SUCCESS);
        setLocalItem("token", access_token);
        setLocalItem("refreshToken", refresh_token);
        setUser(data?.data);
        setIsLogin(true);
        navigate("/");
      }
    } catch (error) {
      errorToast(error?.response?.data?.message);
    } finally {
      setIsLoading(false);
    }
  };

  const handleAccount = (type) => {
    navigate("/register", { state: { type } });
  };

  return (
    <div className="auth-card">
      <div className="flex items-center justify-center mb-5">
        <Logo className="h-16" />
      </div>

      <Form layout="vertical" onFinish={handleSubmit}>
        <Form.Item
          label={t("username")}
          name="username"
          rules={[{ required: true, message: "Please enter your username" }]}
        >
          <Input placeholder={t("enterUsername")} />
        </Form.Item>

        <Form.Item
          label={t("password")}
          name="password"
          rules={[{ required: true, message: "Please enter your password" }]}
        >
          <Input.Password placeholder={t("enterPassword")} />
        </Form.Item>

        <div className="text-right mb-5">
          <Link to="/forgot_password" className="!text-primary">
            {t("forgotPassword")}
          </Link>
        </div>

        <LoadableButton
          type="submit"
          lable={t("login")}
          className="w-full"
          isLoading={isLoading}
          loadingLable={t("loggingIn")}
        />
      </Form>

      <Divider className="!border-theme-border">
        <span>{t("or")}</span>
      </Divider>

      <div className="grid md:grid-cols-2 gap-5">
        <LoadableButton
          lable={t("demoAccount")}
          className="!text-sm"
          onClick={() => handleAccount("demo")}
        />
        <LoadableButton
          lable={t("liveAccount")}
          className="!text-sm"
          onClick={() => handleAccount("live")}
        />
      </div>
    </div>
  );
};

export default Login;
