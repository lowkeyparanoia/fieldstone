import { useMemo, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useCollections, useCreateCollection, useDeleteCollection } from '@/hooks/useApi'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Badge } from '@/components/ui/badge'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import { Plus, Search, MoreVertical, Edit, Trash2, Eye, Database } from 'lucide-react'
import { formatDate } from '@/lib/utils'
import { MOCK_COLLECTIONS } from '@/lib/mock'
import type { Collection } from '@/lib/api'

type Row = Collection & { system?: boolean }

export function CollectionsPage() {
  const navigate = useNavigate()
  const { data } = useCollections()
  const createCollection = useCreateCollection()
  const deleteCollection = useDeleteCollection()

  const base = useMemo<Row[]>(
    () => (data && data.length ? data : MOCK_COLLECTIONS),
    [data]
  )
  // Local overlay so the demo stays interactive even without a live backend.
  const [localRows, setLocalRows] = useState<Row[] | null>(null)
  const rows = localRows ?? base

  const [searchQuery, setSearchQuery] = useState('')
  const [creating, setCreating] = useState(false)
  const [newName, setNewName] = useState('')
  const [collectionToDelete, setCollectionToDelete] = useState<Row | null>(null)

  const filtered = rows.filter((c) =>
    c.name.toLowerCase().includes(searchQuery.toLowerCase())
  )

  const handleCreate = () => {
    const name = newName.trim()
    if (!name) return
    const created: Row = {
      id: 'col_' + Math.random().toString(36).slice(2, 6),
      name,
      schema: [],
      recordCount: 0,
      createdAt: new Date().toISOString(),
      updatedAt: new Date().toISOString(),
    }
    setLocalRows([...rows, created])
    createCollection.mutate({ name, schema: [] }) // best-effort against the API
    setNewName('')
    setCreating(false)
  }

  const handleDelete = () => {
    if (!collectionToDelete) return
    setLocalRows(rows.filter((c) => c.id !== collectionToDelete.id))
    deleteCollection.mutate(collectionToDelete.id) // best-effort against the API
    setCollectionToDelete(null)
  }

  return (
    <div className="space-y-6">
      <div className="flex items-start justify-between gap-4">
        <div>
          <h1 className="text-3xl font-bold tracking-tight">Collections</h1>
          <p className="text-muted-foreground">
            Manage your database collections and schemas
          </p>
        </div>
        <Button onClick={() => setCreating(true)}>
          <Plus className="mr-2 h-4 w-4" />
          New Collection
        </Button>
      </div>

      <Card>
        <CardHeader className="flex flex-row items-center justify-between gap-4 space-y-0">
          <div className="space-y-1.5">
            <CardTitle className="text-2xl">All Collections</CardTitle>
            <CardDescription>{filtered.length} collections in database</CardDescription>
          </div>
          <div className="relative w-[260px]">
            <Search className="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
            <Input
              placeholder="Search collections…"
              className="pl-9"
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
            />
          </div>
        </CardHeader>
        <CardContent>
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Name</TableHead>
                <TableHead>Fields</TableHead>
                <TableHead>Records</TableHead>
                <TableHead>Created</TableHead>
                <TableHead className="w-[50px]" />
              </TableRow>
            </TableHeader>
            <TableBody>
              {filtered.length > 0 ? (
                filtered.map((collection) => (
                  <TableRow key={collection.id}>
                    <TableCell>
                      <div className="flex items-center gap-2.5">
                        <Database className="h-4 w-4 text-muted-foreground" />
                        <button
                          className="font-mono font-medium hover:text-primary"
                          onClick={() => navigate(`/collections/${collection.id}`)}
                        >
                          {collection.name}
                        </button>
                        {collection.system && <Badge variant="secondary">system</Badge>}
                      </div>
                    </TableCell>
                    <TableCell>
                      <Badge variant="secondary">{collection.schema?.length || 0} fields</Badge>
                    </TableCell>
                    <TableCell className="font-mono">
                      {collection.recordCount?.toLocaleString() || 0}
                    </TableCell>
                    <TableCell className="text-muted-foreground">
                      {formatDate(collection.createdAt)}
                    </TableCell>
                    <TableCell>
                      <DropdownMenu>
                        <DropdownMenuTrigger asChild>
                          <Button variant="ghost" size="icon" className="h-8 w-8">
                            <MoreVertical className="h-4 w-4" />
                          </Button>
                        </DropdownMenuTrigger>
                        <DropdownMenuContent align="end">
                          <DropdownMenuItem onClick={() => navigate(`/collections/${collection.id}`)}>
                            <Eye className="mr-2 h-4 w-4" />
                            View Records
                          </DropdownMenuItem>
                          <DropdownMenuItem>
                            <Edit className="mr-2 h-4 w-4" />
                            Edit Schema
                          </DropdownMenuItem>
                          <DropdownMenuSeparator />
                          <DropdownMenuItem
                            className="text-destructive"
                            onClick={() => setCollectionToDelete(collection)}
                          >
                            <Trash2 className="mr-2 h-4 w-4" />
                            Delete
                          </DropdownMenuItem>
                        </DropdownMenuContent>
                      </DropdownMenu>
                    </TableCell>
                  </TableRow>
                ))
              ) : (
                <TableRow>
                  <TableCell colSpan={5} className="py-8 text-center text-muted-foreground">
                    {searchQuery
                      ? 'No collections match your search'
                      : 'No collections found. Create your first collection to get started.'}
                  </TableCell>
                </TableRow>
              )}
            </TableBody>
          </Table>
        </CardContent>
      </Card>

      {/* Create dialog */}
      <Dialog open={creating} onOpenChange={setCreating}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Create Collection</DialogTitle>
            <DialogDescription>
              Create a new collection to store your data. Add fields to define the schema.
            </DialogDescription>
          </DialogHeader>
          <div className="grid gap-2 py-2">
            <label htmlFor="name" className="text-sm font-medium">Collection Name</label>
            <Input
              id="name"
              autoFocus
              placeholder="e.g., products"
              value={newName}
              onChange={(e) => setNewName(e.target.value)}
              onKeyDown={(e) => e.key === 'Enter' && handleCreate()}
            />
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setCreating(false)}>Cancel</Button>
            <Button onClick={handleCreate}>Create Collection</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Delete confirmation */}
      <Dialog open={!!collectionToDelete} onOpenChange={() => setCollectionToDelete(null)}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Delete Collection</DialogTitle>
            <DialogDescription>
              Are you sure you want to delete "{collectionToDelete?.name}"? This action cannot be
              undone and all data will be lost.
            </DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button variant="outline" onClick={() => setCollectionToDelete(null)}>Cancel</Button>
            <Button variant="destructive" onClick={handleDelete}>Delete</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  )
}
