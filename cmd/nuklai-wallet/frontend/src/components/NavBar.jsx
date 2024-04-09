// Copyright (C) 2024, AllianceBlock. All rights reserved.
// See the file LICENSE for licensing terms.

import {
  BankOutlined,
  CheckCircleTwoTone,
  CloseCircleTwoTone,
  ContainerOutlined,
  DashboardOutlined,
  GoldOutlined,
  SendOutlined,
  WalletOutlined,
  WalletTwoTone
} from '@ant-design/icons'
import {
  App,
  Avatar,
  Divider,
  Drawer,
  Layout,
  List,
  Menu,
  Tooltip,
  Typography
} from 'antd'
import { useCallback, useEffect, useState } from 'react'
import { Link as RLink, useLocation } from 'react-router-dom'
import {
  GetAddress,
  GetBalance,
  GetTransactions
} from '../../wailsjs/go/main/App'
import logo from '../assets/images/nuklai-logo.png'
const { Text, Link } = Typography

const NavBar = () => {
  const location = useLocation()
  const { message } = App.useApp()
  const [balances, setBalances] = useState([])
  const [nativeBalance, setNativeBalance] = useState({})
  const [transactions, setTransactions] = useState([])
  const [address, setAddress] = useState('')
  const [drawerOpen, setDrawerOpen] = useState(false)

  // Define the function to fetch data as a useCallback to prevent redefinition on each render
  const fetchData = useCallback(async () => {
    try {
      const [newAddress, newBalances, txs] = await Promise.all([
        GetAddress(),
        GetBalance(),
        GetTransactions()
      ])

      setAddress(newAddress)
      setBalances(newBalances)
      setTransactions(txs.TxInfos)

      for (var bal of newBalances) {
        if (bal.ID == '11111111111111111111111111111111LpoYY') {
          setNativeBalance(bal)
          {
            /* TODO: switch to using context */
          }
          window.HasBalance = bal.Has
          break
        }
      }

      // Handle alerts
      txs.Alerts?.forEach((alert) =>
        message.open({
          icon: <WalletTwoTone />,
          type: alert.Type,
          content: alert.Content
        })
      )
    } catch (error) {
      console.error('Error fetching data:', error)
      message.error('Failed to fetch data. Please try again later.')
    }
  }, [message])

  useEffect(() => {
    fetchData()
    const intervalId = setInterval(fetchData, 5000) // Update every 5 seconds
    return () => clearInterval(intervalId)
  }, [fetchData])

  const formatTimestamp = (timestamp) => {
    const date = new Date(timestamp)
    return date.toLocaleString()
  }

  const shortenAddress = (address) =>
    `${address.slice(0, 6)}...${address.slice(-6)}`

  return (
    <Layout.Header style={{ padding: '0 50px', background: '#fff' }}>
      <div className='logo' style={{ float: 'left' }}>
        <Avatar src={logo} size='large' />
      </div>

      <div
        className='walletInfo'
        style={{ float: 'right', display: 'flex', alignItems: 'center' }}
      >
        <Avatar
          icon={<WalletOutlined />}
          onClick={() => setDrawerOpen(true)}
          style={{ cursor: 'pointer', marginRight: '10px' }}
        />
        <div style={{ display: 'flex', flexDirection: 'column' }}>
          <Tooltip title={address}>
            <Link
              strong
              onClick={() => setDrawerOpen(true)}
              style={{ marginTop: '10px' }}
            >
              {shortenAddress(address)}
            </Link>
          </Tooltip>
          {balances.length > 0 && (
            <Link
              strong
              onClick={() => setDrawerOpen(true)}
              style={{ marginTop: '5px' }}
            >
              {nativeBalance.Str}
            </Link>
          )}
        </div>
      </div>

      <Menu
        mode='horizontal'
        selectedKeys={[location.pathname.slice(1) || 'explorer']}
        items={[
          {
            label: <RLink to={'explorer'}>Explorer</RLink>,
            key: 'explorer',
            icon: <DashboardOutlined />
          },
          {
            label: <RLink to={'faucet'}>Faucet</RLink>,
            key: 'faucet',
            icon: <GoldOutlined />
          },
          {
            label: <RLink to={'mint'}>Mint</RLink>,
            key: 'mint',
            icon: <BankOutlined />
          },
          {
            label: <RLink to={'transfer'}>Transfer</RLink>,
            key: 'transfer',
            icon: <SendOutlined />
          },
          {
            label: <RLink to={'feed'}>Feed</RLink>,
            key: 'feed',
            icon: <ContainerOutlined />
          }
        ]}
      />

      <Drawer
        title={<Text copyable>{address}</Text>}
        width={615}
        placement='right'
        onClose={() => setDrawerOpen(false)}
        open={drawerOpen}
      >
        <Divider orientation='center'>Tokens</Divider>
        <List
          bordered
          dataSource={balances}
          renderItem={(balance) => (
            <List.Item>
              <Text>{balance.Str}</Text>
            </List.Item>
          )}
        />
        <Divider orientation='center'>Transactions</Divider>
        <List
          bordered
          dataSource={transactions}
          renderItem={(item) => (
            <List.Item>
              <div>
                <Text strong>{item.ID} </Text>
                {!item.Success && <CloseCircleTwoTone twoToneColor='#eb2f96' />}
                {item.Success && <CheckCircleTwoTone twoToneColor='#52c41a' />}
              </div>
              <Text strong>Type:</Text> {item.Type}
              <br />
              <Text strong>Timestamp:</Text> {formatTimestamp(item.Timestamp)}
              <br />
              <Text strong>Units:</Text> {item.Units}
              <br />
              <Text strong>Size:</Text> {item.Size}
              <br />
              <Text strong>Summary:</Text> {item.Summary}
              <br />
              <Text strong>Fee:</Text> {item.Fee}
              <br />
              <Text strong>Actor:</Text>{' '}
              <Text copyable>{shortenAddress(item.Actor)}</Text>
            </List.Item>
          )}
        />
      </Drawer>
    </Layout.Header>
  )
}

export default NavBar
