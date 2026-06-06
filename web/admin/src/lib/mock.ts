/**
 * Representative demo data for the admin UI.
 *
 * Shapes are kept in lockstep with the API types in `lib/api.ts`, so the screens
 * render fully even when the Go backend is not running. Pages prefer live
 * TanStack Query data and only fall back to these constants when the API
 * returns nothing — the design stays visible and the data model stays honest.
 */
import type {
  Collection,
  SchemaField,
  CollectionRecord,
  User,
  Activity,
  DashboardStats,
  HealthStatus,
} from './api'

/** A user augmented with admin-panel-only display facets (role/status). */
export interface AdminUser extends User {
  role: 'admin' | 'editor' | 'viewer'
  status: 'active' | 'pending' | 'suspended'
}

const iso = (d: string) => new Date(d).toISOString()

// ---- Schemas (Collection.schema) -------------------------------------------

const SCHEMAS: Record<string, SchemaField[]> = {
  users: [
    { name: 'email', type: 'email', required: true, unique: true },
    { name: 'name', type: 'text', required: false, unique: false },
    { name: 'role', type: 'text', required: false, unique: false },
    { name: 'verified', type: 'boolean', required: false, unique: false },
    { name: 'signup', type: 'date', required: false, unique: false },
  ],
  products: [
    { name: 'name', type: 'text', required: true, unique: false },
    { name: 'price', type: 'number', required: true, unique: false },
    { name: 'sku', type: 'text', required: true, unique: true },
    { name: 'in_stock', type: 'boolean', required: false, unique: false },
    { name: 'category', type: 'relation', required: false, unique: false, options: { collection: 'categories' } },
    { name: 'updated_at', type: 'date', required: false, unique: false },
  ],
  orders: [
    { name: 'order_no', type: 'text', required: true, unique: true },
    { name: 'total', type: 'number', required: true, unique: false },
    { name: 'status', type: 'text', required: true, unique: false },
    { name: 'paid', type: 'boolean', required: false, unique: false },
    { name: 'customer', type: 'relation', required: true, unique: false, options: { collection: 'users' } },
    { name: 'placed_at', type: 'date', required: false, unique: false },
  ],
  posts: [
    { name: 'title', type: 'text', required: true, unique: false },
    { name: 'slug', type: 'text', required: true, unique: true },
    { name: 'published', type: 'boolean', required: false, unique: false },
    { name: 'author', type: 'relation', required: false, unique: false, options: { collection: 'users' } },
    { name: 'homepage', type: 'url', required: false, unique: false },
    { name: 'meta', type: 'json', required: false, unique: false },
    { name: 'date', type: 'date', required: false, unique: false },
  ],
  comments: [
    { name: 'body', type: 'text', required: true, unique: false },
    { name: 'author', type: 'relation', required: false, unique: false, options: { collection: 'users' } },
    { name: 'approved', type: 'boolean', required: false, unique: false },
  ],
  media_assets: [
    { name: 'filename', type: 'text', required: true, unique: false },
    { name: 'url', type: 'url', required: true, unique: false },
    { name: 'size', type: 'number', required: false, unique: false },
  ],
}

// ---- Collections -----------------------------------------------------------

interface MockCollection extends Collection {
  system?: boolean
}

export const MOCK_COLLECTIONS: MockCollection[] = [
  { id: 'col_01', name: 'users', schema: SCHEMAS.users, recordCount: 1284, createdAt: iso('2026-01-02'), updatedAt: iso('2026-01-02'), system: true },
  { id: 'col_02', name: 'products', schema: SCHEMAS.products, recordCount: 8432, createdAt: iso('2026-01-08'), updatedAt: iso('2026-05-30') },
  { id: 'col_03', name: 'orders', schema: SCHEMAS.orders, recordCount: 23981, createdAt: iso('2026-01-08'), updatedAt: iso('2026-06-04') },
  { id: 'col_04', name: 'posts', schema: SCHEMAS.posts, recordCount: 612, createdAt: iso('2026-02-14'), updatedAt: iso('2026-06-04') },
  { id: 'col_05', name: 'comments', schema: SCHEMAS.comments, recordCount: 13402, createdAt: iso('2026-02-14'), updatedAt: iso('2026-06-01') },
  { id: 'col_06', name: 'media_assets', schema: SCHEMAS.media_assets, recordCount: 489, createdAt: iso('2026-03-01'), updatedAt: iso('2026-06-03') },
]

// ---- Records (CollectionRecord.data, keyed by collection name) -------------

const rec = (collectionId: string, id: string, data: Record<string, unknown>): CollectionRecord => ({
  id,
  collectionId,
  data,
  createdAt: iso('2026-05-01'),
  updatedAt: iso('2026-06-01'),
})

export const MOCK_RECORDS: Record<string, CollectionRecord[]> = {
  products: [
    rec('col_02', 'rec_p001', { name: 'Aeron Chair', price: 1395, sku: 'AER-001', in_stock: true, category: 'Furniture', updated_at: '2026-05-30' }),
    rec('col_02', 'rec_p002', { name: 'Standing Desk', price: 680, sku: 'DSK-114', in_stock: true, category: 'Furniture', updated_at: '2026-05-28' }),
    rec('col_02', 'rec_p003', { name: 'Mechanical Keyboard', price: 159, sku: 'KEY-220', in_stock: false, category: 'Peripherals', updated_at: '2026-05-22' }),
    rec('col_02', 'rec_p004', { name: '4K Monitor', price: 549, sku: 'MON-027', in_stock: true, category: 'Peripherals', updated_at: '2026-06-01' }),
    rec('col_02', 'rec_p005', { name: 'USB-C Hub', price: 49, sku: 'HUB-009', in_stock: true, category: 'Accessories', updated_at: '2026-06-03' }),
    rec('col_02', 'rec_p006', { name: 'Laptop Stand', price: 38, sku: 'STD-061', in_stock: false, category: 'Accessories', updated_at: '2026-05-19' }),
  ],
  orders: [
    rec('col_03', 'rec_o001', { order_no: 'ORD-10241', total: 1444, status: 'shipped', paid: true, customer: 'usr_8fa2c1', placed_at: '2026-06-02' }),
    rec('col_03', 'rec_o002', { order_no: 'ORD-10242', total: 208, status: 'processing', paid: true, customer: 'usr_71bce4', placed_at: '2026-06-03' }),
    rec('col_03', 'rec_o003', { order_no: 'ORD-10243', total: 549, status: 'pending', paid: false, customer: 'usr_2a90fd', placed_at: '2026-06-04' }),
    rec('col_03', 'rec_o004', { order_no: 'ORD-10244', total: 87, status: 'shipped', paid: true, customer: 'usr_9d44e0', placed_at: '2026-06-04' }),
    rec('col_03', 'rec_o005', { order_no: 'ORD-10245', total: 1944, status: 'cancelled', paid: false, customer: 'usr_5c12ab', placed_at: '2026-06-05' }),
  ],
  users: [
    rec('col_01', 'rec_u001', { email: 'deep@fieldstone.io', name: 'Deep', role: 'admin', verified: true, signup: '2026-01-02' }),
    rec('col_01', 'rec_u002', { email: 'rao@championsmail.com', name: 'Rao', role: 'editor', verified: true, signup: '2026-01-09' }),
    rec('col_01', 'rec_u003', { email: 'mira@acme.dev', name: 'Mira', role: 'viewer', verified: true, signup: '2026-02-14' }),
    rec('col_01', 'rec_u004', { email: 'jonas@vega.app', name: 'Jonas', role: 'editor', verified: false, signup: '2026-03-01' }),
  ],
  posts: [
    rec('col_04', 'rec_b001', { title: 'Shipping realtime to production', slug: 'realtime-prod', published: true, author: 'usr_8fa2c1', homepage: 'https://fieldstone.io/blog', meta: { reads: 1204 }, date: '2026-05-20' }),
    rec('col_04', 'rec_b002', { title: 'Designing the schema editor', slug: 'schema-editor', published: true, author: 'usr_71bce4', homepage: 'https://fieldstone.io/blog', meta: { reads: 642 }, date: '2026-05-28' }),
    rec('col_04', 'rec_b003', { title: 'Multi-tenancy patterns', slug: 'multi-tenancy', published: false, author: 'usr_8fa2c1', homepage: '', meta: { draft: true }, date: '2026-06-04' }),
  ],
  comments: [
    rec('col_05', 'rec_c001', { body: 'This unblocked our migration — thanks!', author: 'usr_2a90fd', approved: true }),
    rec('col_05', 'rec_c002', { body: 'Any plans for row-level security?', author: 'usr_9d44e0', approved: true }),
    rec('col_05', 'rec_c003', { body: 'first', author: 'usr_b3f871', approved: false }),
  ],
  media_assets: [
    rec('col_06', 'rec_m001', { filename: 'hero.png', url: 'https://cdn.fieldstone.io/hero.png', size: 13057 }),
    rec('col_06', 'rec_m002', { filename: 'og-card.jpg', url: 'https://cdn.fieldstone.io/og-card.jpg', size: 88210 }),
  ],
}

// ---- Users -----------------------------------------------------------------

export const MOCK_USERS: AdminUser[] = [
  { id: 'usr_8fa2c1', email: 'deep@fieldstone.io', role: 'admin', verified: true, status: 'active', lastLoginAt: iso('2026-06-06T11:58:00'), createdAt: iso('2026-01-02'), updatedAt: iso('2026-06-06') },
  { id: 'usr_71bce4', email: 'rao@championsmail.com', role: 'editor', verified: true, status: 'active', lastLoginAt: iso('2026-06-06T11:46:00'), createdAt: iso('2026-01-09'), updatedAt: iso('2026-06-06') },
  { id: 'usr_2a90fd', email: 'mira@acme.dev', role: 'viewer', verified: true, status: 'active', lastLoginAt: iso('2026-06-06T11:00:00'), createdAt: iso('2026-02-14'), updatedAt: iso('2026-06-06') },
  { id: 'usr_5c12ab', email: 'jonas@vega.app', role: 'editor', verified: false, status: 'pending', lastLoginAt: iso('2026-06-06T09:00:00'), createdAt: iso('2026-03-01'), updatedAt: iso('2026-06-06') },
  { id: 'usr_9d44e0', email: 'sara@northwind.co', role: 'viewer', verified: true, status: 'active', lastLoginAt: iso('2026-06-05T12:00:00'), createdAt: iso('2026-03-18'), updatedAt: iso('2026-06-05') },
  { id: 'usr_b3f871', email: 'lee@orbit.sh', role: 'viewer', verified: false, status: 'suspended', lastLoginAt: iso('2026-06-02T12:00:00'), createdAt: iso('2026-04-02'), updatedAt: iso('2026-06-02') },
]

/** Map a live API user onto the admin display shape with sensible defaults. */
export function toAdminUser(u: User): AdminUser {
  return {
    ...u,
    role: 'viewer',
    status: u.verified ? 'active' : 'pending',
  }
}

// ---- Dashboard -------------------------------------------------------------

export const MOCK_STATS: DashboardStats = {
  totalRecords: 48210,
  totalUsers: 1284,
  totalCollections: 6,
  requestsPerMinute: 342,
  storageUsed: 2.4 * 1024 * 1024 * 1024,
  storageLimit: 10 * 1024 * 1024 * 1024,
  activeUsers: 87,
}

export const MOCK_HEALTH: HealthStatus = {
  status: 'ok',
  timestamp: iso('2026-06-06T11:59:30'),
  services: { database: 'healthy', api: 'healthy', cache: 'unhealthy' },
}

/** Request volume over the last 24h — the dashboard area chart series. */
export const MOCK_CHART = [
  { time: '00:00', requests: 120 },
  { time: '04:00', requests: 80 },
  { time: '08:00', requests: 340 },
  { time: '12:00', requests: 520 },
  { time: '16:00', requests: 680 },
  { time: '20:00', requests: 420 },
  { time: '23:59', requests: 280 },
]

// ---- Activity --------------------------------------------------------------

export const MOCK_ACTIVITIES: Activity[] = [
  { id: '1', type: 'record_created', message: 'New record created in products', userId: 'usr_8fa2c1', createdAt: iso('2026-06-06T11:58:00') },
  { id: '2', type: 'collection_modified', message: 'Collection orders schema modified', userId: 'usr_71bce4', createdAt: iso('2026-06-06T11:38:00') },
  { id: '3', type: 'record_updated', message: 'Record updated in users', userId: 'usr_8fa2c1', createdAt: iso('2026-06-06T11:12:00') },
  { id: '4', type: 'backup_completed', message: 'Scheduled database backup completed', createdAt: iso('2026-06-06T11:00:00') },
  { id: '5', type: 'webhook_triggered', message: 'Webhook order.created delivered (200)', createdAt: iso('2026-06-06T10:55:00') },
  { id: '6', type: 'user_created', message: 'New user registered: jonas@vega.app', createdAt: iso('2026-06-05T15:20:00') },
  { id: '7', type: 'record_deleted', message: 'Deleted 3 records from comments', userId: 'usr_71bce4', createdAt: iso('2026-06-05T14:02:00') },
  { id: '8', type: 'collection_modified', message: 'Created collection media_assets', userId: 'usr_8fa2c1', createdAt: iso('2026-06-05T09:30:00') },
]
