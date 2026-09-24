import dashboard from './dashboard'
import misc from './misc'
import upstreamUpdate from './upstreamUpdate'

export default {
  ...upstreamUpdate,
  ...dashboard,
  ...misc,
}
