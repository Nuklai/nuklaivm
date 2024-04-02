// Copyright (C) 2024, AllianceBlock. All rights reserved.
// See the file LICENSE for licensing terms.

import { App as AApp, FloatButton, Layout, Row, Typography } from 'antd'
import { useEffect, useState } from 'react'
import { Outlet } from 'react-router-dom'
import { GetCommitHash, OpenLink } from '../wailsjs/go/main/App'
import logo from './assets/images/nuklai-footer.png'
import NavBar from './components/NavBar'

const { Text } = Typography
const { Content } = Layout

const App = () => {
  const [commit, setCommit] = useState('')
  useEffect(() => {
    const getCommit = async () => {
      const c = await GetCommitHash()
      setCommit(c)
    }
    getCommit()
  }, [])
  return (
    <AApp>
      <Layout
        style={{
          minHeight: '95vh'
        }}
      >
        <NavBar />
        <Layout className='site-layout'>
          <Content
            style={{
              background: 'white',
              padding: '0 50px'
            }}
          >
            <div
              style={{
                padding: 24
              }}
            >
              <Outlet />
              <FloatButton.BackTop />
            </div>
          </Content>
        </Layout>
        <Row justify='center' style={{ background: 'white' }}>
          <a
            onClick={() => {
              OpenLink('https://github.com/ava-labs/hypersdk')
            }}
          >
            <img src={logo} style={{ width: '300px' }} />
          </a>
        </Row>
        <Row justify='center' style={{ background: 'white' }}>
          <Text type='secondary'>Commit: {commit}</Text>
        </Row>
      </Layout>
    </AApp>
  )
}

export default App
