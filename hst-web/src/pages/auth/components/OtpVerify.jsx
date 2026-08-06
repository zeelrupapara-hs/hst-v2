// import { useEffect, useState, useMemo } from "react";
import { useTranslation } from "react-i18next";
import { Form, Input } from "antd";
import LoadableButton from "../../../components/button/LoadableButton";
// import Spinner from "../../../components/loader/Spinner";
// import { successToast, errorToast } from "../../../utils/utils";
// import { resendOtp } from "../../../api/request/auth";

const OtpVerify = ({ email, handleSubmit, isLoading }) => {
  const { t } = useTranslation();
  // const [isResendLoading, setIsResendLoading] = useState(false);
  // const [timer, setTimer] = useState(59);

  // useEffect(() => {
  //   const interval = setInterval(() => {
  //     setTimer((current) => (current !== 0 ? current - 1 : 0));
  //   }, 1000);

  //   return () => clearInterval(interval);
  // }, []);

  // const formattedTime = useMemo(() => {
  //   const minutes = Math.floor(timer / 60);
  //   const remainingSeconds = timer % 60;
  //   return `${String(minutes).padStart(2, "0")}:${String(
  //     remainingSeconds
  //   ).padStart(2, "0")}`;
  // }, [timer]);

  // const handleResend = async () => {
  //   setIsResendLoading(true);
  //   try {
  //     const { data } = await resendOtp({ email });
  //     if (data?.success) {
  //       successToast(data?.message);
  //       setTimer(59);
  //     }
  //   } catch (error) {
  //     errorToast(error?.response?.data?.message);
  //   } finally {
  //     setIsResendLoading(false);
  //   }
  // };

  return (
    <Form layout="vertical" onFinish={handleSubmit}>
      <Form.Item
        name="otp"
        rules={[
          { required: true, message: "Please enter the OTP" },
          { pattern: /^\d{5}$/, message: "OTP must be exactly 5 digits" },
        ]}
      >
        <Input.OTP length={5} autoFocus />
      </Form.Item>

      {/* <div className="text-center my-5">
        {timer === 0 ? (
          <p>
            Don't receive any code?{" "}
            <span
              onClick={handleResend}
              className="text-primary cursor-pointer ml-1"
            >
              {isResendLoading ? <Spinner size="small" /> : "Resend"}
            </span>
          </p>
        ) : (
          <p>
            Resend OTP in <span className="text-primary">{formattedTime}</span>
          </p>
        )}
      </div> */}

      <LoadableButton
        type="submit"
        lable={t("verifyOtp")}
        isLoading={isLoading}
        loadingLable={t("verifyingOtp")}
        className="w-full"
      />
    </Form>
  );
};

export default OtpVerify;
