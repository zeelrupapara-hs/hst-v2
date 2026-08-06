import { useEffect, useMemo, useState } from "react";
import { Link, useLocation, useNavigate } from "react-router-dom";
import { Trans, useTranslation } from "react-i18next";
import { Form, Input, Select } from "antd";
import countryList from "country-list";
import LoadableButton from "../../components/button/LoadableButton";
import { register } from "../../api/request/auth";
import { errorToast, successToast } from "../../utils/utils";
import MESSAGES from "../../utils/messages";
import Logo from "../../components/common/Logo";

const { Option } = Select;

const Register = () => {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const { state } = useLocation();
  const [type, setType] = useState("demo");
  const [email, setEmail] = useState("");
  const [step, setStep] = useState(1);
  const [isLoading, setIsLoading] = useState(false);
  const countries = useMemo(() => countryList.getData(), []);
  const amounts = useMemo(() => [3000, 5000, 10000, 25000, 50000, 100000], []);

  useEffect(() => {
    if (state?.type) setType(state?.type);
  }, [state?.type]);

  const handleSubmit = async (values) => {
    setIsLoading(true);
    try {
      const body = { ...values, trade_type: type === "demo" ? 1 : 0 };
      const { data } = await register(body);
      if (data?.success) {
        successToast(data?.message || MESSAGES.REGISTER_SUCCESS);
        setEmail(values?.email);
        setStep(2);
      }
    } catch (error) {
      errorToast(error?.response?.data?.message);
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div className="auth-card !max-w-xl">
      <div className="flex items-center justify-center mb-5">
        <Logo className="h-16" />
      </div>

      {step === 1 && (
        <div>
          <Form layout="vertical" onFinish={handleSubmit}>
            <div className="grid md:grid-cols-2 md:gap-4">
              <Form.Item
                label={t("firstName")}
                name="first_name"
                rules={[{ required: true, message: "Please enter first name" }]}
              >
                <Input placeholder={t("enterFirstName")} />
              </Form.Item>

              <Form.Item
                label={t("lastName")}
                name="last_name"
                rules={[{ required: true, message: "Please enter last name" }]}
              >
                <Input placeholder={t("enterLastName")} />
              </Form.Item>
            </div>

            <div className="grid md:grid-cols-2 md:gap-4">
              <Form.Item
                label={t("username")}
                name="username"
                rules={[
                  { required: true, message: "Please enter your username" },
                ]}
              >
                <Input placeholder={t("enterUsername")} />
              </Form.Item>

              <Form.Item
                label={t("email")}
                name="email"
                rules={[
                  { required: true, message: "Please enter your email" },
                  { type: "email", message: "Please enter valid email" },
                ]}
              >
                <Input placeholder={t("enterEmail")} />
              </Form.Item>
            </div>

            <div className="grid md:grid-cols-2 md:gap-4">
              <Form.Item label={t("country")} name="country">
                <Select
                  placeholder={t("selectCountry")}
                  showSearch
                  optionFilterProp="children"
                  filterOption={(input, option) =>
                    option?.children
                      ?.toLowerCase()
                      .includes(input?.toLowerCase())
                  }
                >
                  {countries?.map((item, index) => (
                    <Option key={index} value={item?.name}>
                      {item?.name}
                    </Option>
                  ))}
                </Select>
              </Form.Item>

              <Form.Item
                label={t("phone")}
                name="phone"
                rules={[
                  { required: true, message: "Please enter phone number" },
                  {
                    pattern: /^[0-9]{10}$/,
                    message: "Phone number must be 10 digits",
                  },
                ]}
              >
                <Input type="number" placeholder={t("enterPhone")} />
              </Form.Item>
            </div>

            <Form.Item label={t("address")} name="address">
              <Input placeholder={t("enterAddress")} />
            </Form.Item>

            {type === "demo" && (
              <Form.Item
                label={t("deposit")}
                name="deposit"
                initialValue={100000}
              >
                <Select placeholder={t("selectDeposit")} prefix="$">
                  {amounts?.map((value, index) => (
                    <Option key={index} value={value}>
                      {value}
                    </Option>
                  ))}
                </Select>
              </Form.Item>
            )}

            <LoadableButton
              type="submit"
              lable={t("openAccount", { type: t(type) })}
              className="w-full capitalize"
              isLoading={isLoading}
              loadingLable={t("openingAccount", { type: t(type) })}
            />
          </Form>

          <p className="text-center mt-5">
            {t("alreadyHaveAccount")}{" "}
            <Link to="/login" className="!text-primary">
              {t("login")}
            </Link>
          </p>
        </div>
      )}

      {step === 2 && (
        <div className="text-center space-y-5">
          <h2 className="text-3xl font-semibold text-green-500 capitalize">
            {t("congratulations")}!
          </h2>

          <p className="text-base text-gray">
            <Trans
              i18nKey="accountSuccessMessage"
              values={{ type: t(type), email }}
              components={{ 1: <span className="font-semibold" /> }}
            />
          </p>

          <LoadableButton
            lable={t("backToLogin")}
            className="!text-sm"
            onClick={() => navigate("/login")}
          />
        </div>
      )}
    </div>
  );
};

export default Register;
