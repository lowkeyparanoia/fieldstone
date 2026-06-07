import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { Table2, Plus } from 'lucide-react'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogTrigger } from '@/components/ui/dialog'
import { gatewayApi } from '@/lib/api'

export function GatewayTablesPage() {
  const queryClient = useQueryClient()
  const [selectedTable, setSelectedTable] = useState<string | null>(null)
  const [filterCol, setFilterCol] = useState('')
  const [filterVal, setFilterVal] = useState('')
  const [newRow, setNewRow] = useState<Record<string, string>>({})

  const { data: tablesData } = useQuery({
    queryKey: ['tables'],
    queryFn: async () => {
      return { tables: ['users', 'companies', 'prospects', 'campaigns', 'engagements'] }
    },
  })

  const { data: rowsData, isLoading } = useQuery({
    queryKey: ['rows', selectedTable, filterCol, filterVal],
    queryFn: async () => {
      if (!selectedTable) return { data: [] }
      const params: Record<string, string> = {}
      if (filterCol && filterVal) {
        params[`eq.${filterCol}`] = filterVal
      }
      const res = await gatewayApi.query(selectedTable, params)
      return res.data
    },
    enabled: !!selectedTable,
  })

  const insertMutation = useMutation({
    mutationFn: () => {
      if (!selectedTable) throw new Error('No table selected')
      return gatewayApi.insert(selectedTable, newRow)
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['rows', selectedTable] })
      setNewRow({})
    },
  })

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-3xl font-bold flex items-center gap-2">
          <Table2 className="h-8 w-8" /> Tables (PostgREST Gateway)
        </h1>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-4 gap-6">
        <Card className="lg:col-span-1">
          <CardHeader>
            <CardTitle>Tables</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-2">
              {tablesData?.tables.map((table: string) => (
                <Button
                  key={table}
                  variant={selectedTable === table ? 'default' : 'ghost'}
                  className="w-full justify-start"
                  onClick={() => setSelectedTable(table)}
                >
                  <Table2 className="mr-2 h-4 w-4" />
                  {table}
                </Button>
              ))}
            </div>
          </CardContent>
        </Card>

        <Card className="lg:col-span-3">
          <CardHeader>
            <CardTitle className="flex items-center justify-between">
              <span>{selectedTable || 'Select a table'}</span>
              {selectedTable && (
                <div className="flex items-center gap-2">
                  <Input
                    placeholder="Column"
                    className="w-32"
                    value={filterCol}
                    onChange={(e) => setFilterCol(e.target.value)}
                  />
                  <Input
                    placeholder="Value"
                    className="w-32"
                    value={filterVal}
                    onChange={(e) => setFilterVal(e.target.value)}
                  />
                </div>
              )}
            </CardTitle>
          </CardHeader>
          <CardContent>
            {!selectedTable ? (
              <div className="text-sm text-muted-foreground">Select a table to browse</div>
            ) : isLoading ? (
              <div className="text-sm text-muted-foreground">Loading...</div>
            ) : (
              <div className="space-y-4">
                <div className="border rounded-md overflow-auto max-h-[400px]">
                  <table className="w-full text-sm">
                    <thead className="bg-muted">
                      <tr>
                        {!!rowsData?.data?.[0] && Object.keys(rowsData.data[0] as Record<string, unknown>).map((col) => (
                          <th key={col} className="px-4 py-2 text-left font-medium">{col}</th>
                        ))}
                      </tr>
                    </thead>
                    <tbody>
                      {(rowsData?.data || []).map((row: any, idx: number) => (
                        <tr key={idx} className="border-t hover:bg-muted/50">
                          {Object.values(row).map((val: any, vidx: number) => (
                            <td key={vidx} className="px-4 py-2">
                              {typeof val === 'object' ? JSON.stringify(val) : String(val)}
                            </td>
                          ))}
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>

                <Dialog>
                  <DialogTrigger asChild>
                    <Button><Plus className="mr-2 h-4 w-4" /> Insert Row</Button>
                  </DialogTrigger>
                  <DialogContent>
                    <DialogHeader>
                      <DialogTitle>Insert into {selectedTable}</DialogTitle>
                    </DialogHeader>
                    <div className="space-y-4">
                      <Input
                        placeholder='JSON: {"name": "value"}'
                        value={JSON.stringify(newRow)}
                        onChange={(e) => {
                          try {
                            setNewRow(JSON.parse(e.target.value))
                          } catch {}
                        }}
                      />
                      <Button onClick={() => insertMutation.mutate()}>Insert</Button>
                    </div>
                  </DialogContent>
                </Dialog>
              </div>
            )}
          </CardContent>
        </Card>
      </div>
    </div>
  )
}
