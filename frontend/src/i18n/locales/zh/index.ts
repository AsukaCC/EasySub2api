import core from './core'
import app from './app'
import channelMonitorV2 from './channelMonitorV2'
import admin from './admin'
import adminAccessKeys from './adminAccessKeys'

export default {
  ...core,
  ...app,
  ...channelMonitorV2,
  admin,
  adminAccessKeys,
}
