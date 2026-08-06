import React from 'react'
import { Spin } from 'antd';
import { LoadingOutlined } from '@ant-design/icons';

const Spinner = ({ size = "" }) => {
  return (
    <Spin indicator={<LoadingOutlined spin />} size={size} />
  )
}

export default Spinner
