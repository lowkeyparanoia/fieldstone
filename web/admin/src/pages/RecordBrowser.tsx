import { useMemo, useState } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import { useCollections, useRecords, useCreateRecord } from '@/hooks/useApi'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Badge } from '@/components/ui/badge'
import { Label } from '@/components/ui/label'
import { Switch } from '@/components/ui/switch'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import {
  ChevronLeft,
  ChevronRight,
  Plus,
  Search,
  Save,
  Trash2,
  X,
  Type,
  Hash,
  ToggleLeft,
  Calendar,
  Mail,
  Link,
  Braces,
  Link2,
  type LucideIcon,
} from 'lucide-react'
import { MOCK_COLLECTIONS, MOCK_RECORDS } from '@/lib/mock'
import type { Collection, SchemaField, CollectionRecord } from '@/lib/api'

const TYPE_META: Record<SchemaField['type'], { icon: LucideIcon; label: string }> = {
  text: { icon: Type, label: 'text' },
  number: { icon: Hash, label: 'number' },
  boolean: { icon: ToggleLeft, label: 'boolean' },
  date: { icon: Calendar, label: 'date' },
  email: { icon: Mail, label: 'email' },
  url: { icon: Link, label: 'url' },
  json: { icon: Braces, label: 'json' },
  relation: { icon: Link2, label: 'relation' },
}

function relLabel(collection: string | undefined, id: unknown): string {
  if (!collection) return String(id)
  const recs = MOCK_RECORDS[collection] || []
  const r = recs.find((x) => x.id === id)
  if (!r) return String(id)
  const firstText = Object.values(r.data).find((v) => typeof v === 'string')
  return (firstText as string) || String(id)
}

function Cell({ field, value }: { field: SchemaField; value: unknown }) {
  if (value === undefined || value === null || value === '') {
    return <span className="text-muted-foreground">—</span>
  }
  switch (field.type) {
    case 'boolean':
      return <Badge variant={value ? 'success' : 'secondary'}>{String(value)}</Badge>
    case 'number':
      return (
        <span className="font-mono">
          {field.name === 'price' || field.name === 'total' ? '$' : ''}
          {(value as number).toLocaleString()}
        </span>
      )
    case 'date':
      return <span className="font-mono text-muted-foreground">{String(value)}</span>
    case 'email':
      return <span className="font-mono">{String(value)}</span>
    case 'url':
      return (
        <span className="inline-block max-w-[180px] truncate align-bottom font-mono text-muted-foreground">
          {String(value)}
        </span>
      )
    case 'json':
      return (
        <span className="font-mono text-muted-foreground">
          {'{ '}
          {Object.keys(value as object).length} keys
          {' }'}
        </span>
      )
    case 'relation':
      return (
        <span className="inline-flex items-center gap-1.5 font-mono text-primary">
          <Link2 className="h-3 w-3" />
          {relLabel(field.options?.collection as string | undefined, value)}
        </span>
      )
    default:
      return <span className="font-medium">{String(value)}</span>
  }
}

type EditorTarget = CollectionRecord | 'new' | null

function RecordEditor({
  collectionName,
  schema,
  record,
  onClose,
  onSave,
  onDelete,
}: {
  collectionName: string
  schema: SchemaField[]
  record: CollectionRecord | null
  onClose: () => void
  onSave: (data: Record<string, unknown>, id?: string) => void
  onDelete: (id: string) => void
}) {
  const isNew = !record
  const [form, setForm] = useState<Record<string, unknown>>(() => {
    const base: Record<string, unknown> = {}
    schema.forEach((f) => {
      base[f.name] = record ? record.data[f.name] : f.type === 'boolean' ? false : ''
    })
    return base
  })
  const upd = (k: string, v: unknown) => setForm((f) => ({ ...f, [k]: v }))

  const renderField = (f: SchemaField) => {
    const v = form[f.name]
    if (f.type === 'boolean') {
      return (
        <div className="flex items-center justify-between">
          <Label>{f.name}</Label>
          <Switch checked={!!v} onCheckedChange={(c) => upd(f.name, c)} />
        </div>
      )
    }
    if (f.type === 'relation') {
      const target = f.options?.collection as string | undefined
      const opts = (target ? MOCK_RECORDS[target] || [] : []).map((r) => ({
        value: r.id,
        label: `${relLabel(target, r.id)} · ${r.id}`,
      }))
      return (
        <div className="flex flex-col gap-2">
          <Label>
            {f.name}{' '}
            <span className="font-mono font-normal text-muted-foreground">→ {target}</span>
          </Label>
          <Select value={(v as string) || undefined} onValueChange={(val) => upd(f.name, val)}>
            <SelectTrigger>
              <SelectValue placeholder={`Select ${target}…`} />
            </SelectTrigger>
            <SelectContent>
              {opts.map((o) => (
                <SelectItem key={o.value} value={o.value}>
                  {o.label}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>
      )
    }
    if (f.type === 'json') {
      return (
        <div className="flex flex-col gap-2">
          <Label>{f.name}</Label>
          <textarea
            rows={4}
            className="resize-y rounded-md border border-input bg-background p-2.5 font-mono text-[13px] outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 focus-visible:ring-offset-background"
            value={typeof v === 'object' && v !== null ? JSON.stringify(v, null, 2) : (v as string) || ''}
            onChange={(e) => upd(f.name, e.target.value)}
          />
        </div>
      )
    }
    const inputType =
      f.type === 'number' ? 'number' : f.type === 'date' ? 'date' : f.type === 'email' ? 'email' : 'text'
    return (
      <div className="flex flex-col gap-2">
        <Label>{f.name}</Label>
        <Input
          type={inputType}
          value={v == null ? '' : (v as string | number)}
          placeholder={f.type}
          onChange={(e) => upd(f.name, e.target.value)}
        />
      </div>
    )
  }

  return (
    <div
      className="fixed inset-0 z-50 flex justify-end bg-black/60 animate-fade-in"
      onClick={onClose}
    >
      <div
        className="flex h-full w-[460px] max-w-[92vw] flex-col border-l bg-card shadow-lg animate-slide-in-right"
        onClick={(e) => e.stopPropagation()}
      >
        <div className="flex items-center justify-between border-b px-6 py-5">
          <div>
            <h2 className="text-lg font-semibold">{isNew ? 'New record' : 'Edit record'}</h2>
            <p className="mt-0.5 font-mono text-xs text-muted-foreground">
              {isNew ? collectionName : record!.id}
            </p>
          </div>
          <Button variant="ghost" size="icon" onClick={onClose}>
            <X className="h-[18px] w-[18px]" />
          </Button>
        </div>

        <div className="flex flex-1 flex-col gap-[18px] overflow-y-auto p-6">
          {schema.map((f) => {
            const Meta = TYPE_META[f.type]
            return (
              <div key={f.name}>
                {renderField(f)}
                <div className="mt-1.5 flex gap-1.5">
                  <Badge variant="outline" className="gap-1">
                    <Meta.icon className="h-3 w-3" />
                    {Meta.label}
                  </Badge>
                  {f.required && <Badge variant="secondary">required</Badge>}
                  {f.unique && <Badge variant="secondary">unique</Badge>}
                </div>
              </div>
            )
          })}
        </div>

        <div className="flex items-center justify-between border-t px-6 py-4">
          {!isNew ? (
            <Button variant="outline" className="text-destructive" onClick={() => onDelete(record!.id)}>
              <Trash2 className="mr-2 h-4 w-4" />
              Delete
            </Button>
          ) : (
            <span />
          )}
          <div className="flex gap-2">
            <Button variant="outline" onClick={onClose}>Cancel</Button>
            <Button onClick={() => onSave(form, record?.id)}>
              <Save className="mr-2 h-4 w-4" />
              {isNew ? 'Create' : 'Save'}
            </Button>
          </div>
        </div>
      </div>
    </div>
  )
}

export function RecordBrowserPage() {
  const { id = '' } = useParams()
  const navigate = useNavigate()
  const { data: collectionsData } = useCollections()

  const collections: Collection[] = collectionsData && collectionsData.length ? collectionsData : MOCK_COLLECTIONS
  const collection = collections.find((c) => c.id === id)

  const { data: recordsData } = useRecords(id)
  const createRecord = useCreateRecord(id)

  const seeded = useMemo<CollectionRecord[]>(() => {
    if (recordsData?.items && recordsData.items.length) return recordsData.items
    return collection ? MOCK_RECORDS[collection.name] || [] : []
  }, [recordsData, collection])

  const [localRecords, setLocalRecords] = useState<CollectionRecord[] | null>(null)
  const records = localRecords ?? seeded

  const [query, setQuery] = useState('')
  const [editing, setEditing] = useState<EditorTarget>(null)

  if (!collection) {
    return (
      <div className="space-y-6">
        <Button variant="ghost" className="px-0 text-muted-foreground" onClick={() => navigate('/collections')}>
          <ChevronLeft className="mr-1.5 h-4 w-4" />
          Collections
        </Button>
        <Card>
          <CardContent className="py-16 text-center text-muted-foreground">
            Collection not found.
          </CardContent>
        </Card>
      </div>
    )
  }

  const schema = collection.schema || []
  const filtered = records.filter(
    (r) =>
      JSON.stringify(r.data).toLowerCase().includes(query.toLowerCase()) ||
      r.id.toLowerCase().includes(query.toLowerCase())
  )

  const save = (data: Record<string, unknown>, recId?: string) => {
    if (recId) {
      setLocalRecords(records.map((r) => (r.id === recId ? { ...r, data } : r)))
    } else {
      const created: CollectionRecord = {
        id: 'rec_' + Math.random().toString(36).slice(2, 7),
        collectionId: collection.id,
        data,
        createdAt: new Date().toISOString(),
        updatedAt: new Date().toISOString(),
      }
      setLocalRecords([...records, created])
      createRecord.mutate(data) // best-effort against the API
    }
    setEditing(null)
  }
  const del = (recId: string) => {
    setLocalRecords(records.filter((r) => r.id !== recId))
    setEditing(null)
  }

  return (
    <div className="space-y-6">
      {/* Breadcrumb + header */}
      <div>
        <button
          className="mb-2.5 inline-flex items-center gap-1.5 text-[13px] text-muted-foreground hover:text-foreground"
          onClick={() => navigate('/collections')}
        >
          <ChevronLeft className="h-[15px] w-[15px]" />
          Collections
        </button>
        <div className="flex items-start justify-between gap-4">
          <div>
            <h1 className="font-mono text-3xl font-bold tracking-tight">{collection.name}</h1>
            <p className="mt-1.5 text-muted-foreground">
              {records.length} records · {schema.length} fields
            </p>
          </div>
          <Button onClick={() => setEditing('new')}>
            <Plus className="mr-2 h-4 w-4" />
            New record
          </Button>
        </div>
      </div>

      {/* Schema strip */}
      <div className="flex flex-wrap gap-2">
        {schema.map((f) => {
          const Meta = TYPE_META[f.type]
          return (
            <span
              key={f.name}
              className="inline-flex items-center gap-1.5 rounded-md border bg-card px-2.5 py-1.5 text-xs"
            >
              <Meta.icon className="h-[13px] w-[13px] text-muted-foreground" />
              <span className="font-mono">{f.name}</span>
              <span className="font-mono text-[11px] text-muted-foreground">
                {f.type}
                {f.required ? '*' : ''}
              </span>
            </span>
          )
        })}
      </div>

      <Card>
        <CardHeader className="flex flex-row items-center justify-between gap-4 space-y-0">
          <div className="space-y-1.5">
            <CardTitle className="text-lg">Records</CardTitle>
            <CardDescription>Browse and edit rows in this collection</CardDescription>
          </div>
          <div className="relative w-[260px]">
            <Search className="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
            <Input
              placeholder="Filter records…"
              className="pl-9"
              value={query}
              onChange={(e) => setQuery(e.target.value)}
            />
          </div>
        </CardHeader>
        <CardContent className="p-0">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead className="whitespace-nowrap pl-6">id</TableHead>
                {schema.map((f) => (
                  <TableHead key={f.name} className="whitespace-nowrap">
                    {f.name}
                  </TableHead>
                ))}
                <TableHead className="w-[44px]" />
              </TableRow>
            </TableHeader>
            <TableBody>
              {filtered.length > 0 ? (
                filtered.map((r) => (
                  <TableRow
                    key={r.id}
                    className="cursor-pointer"
                    onClick={() => setEditing(r)}
                  >
                    <TableCell className="whitespace-nowrap pl-6 font-mono text-muted-foreground">
                      {r.id}
                    </TableCell>
                    {schema.map((f) => (
                      <TableCell key={f.name} className="whitespace-nowrap">
                        <Cell field={f} value={r.data[f.name]} />
                      </TableCell>
                    ))}
                    <TableCell>
                      <ChevronRight className="h-4 w-4 text-muted-foreground" />
                    </TableCell>
                  </TableRow>
                ))
              ) : (
                <TableRow>
                  <TableCell
                    colSpan={schema.length + 2}
                    className="py-8 text-center text-muted-foreground"
                  >
                    No records match your filter
                  </TableCell>
                </TableRow>
              )}
            </TableBody>
          </Table>
        </CardContent>
      </Card>

      {editing && (
        <RecordEditor
          collectionName={collection.name}
          schema={schema}
          record={editing === 'new' ? null : editing}
          onClose={() => setEditing(null)}
          onSave={save}
          onDelete={del}
        />
      )}
    </div>
  )
}
