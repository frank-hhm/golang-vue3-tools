<template>
  <layout-body>
    <div class="mysql-page">
      <!-- 左侧：连接 + 对象树 -->
      <div class="left-panel">
        <div class="lp-block">
          <div class="lp-head">
            <span class="lp-title">连接</span>
            <a-button size="mini" type="text" @click="openAddModal()"><template
                #icon><icon-plus /></template></a-button>
          </div>
          <div class="conn-list">
            <div v-for="c in savedConns" :key="c.id" class="conn-item" :class="{ active: c.id === st.activeConnId }"
              @click="openConnection(c)">
              <span class="led" :class="{ on: isOpen(c.id) }"></span>
              <span class="conn-name">{{ c.name || (c.host + ':' + c.port) }}</span>
              <a-dropdown trigger="hover" @select="(v: any) => onConnAction(v, c)">
                <icon-more class="conn-more" @click.stop />
                <template #content>
                  <a-doption value="open"><icon-play-arrow /> 打开连接</a-doption>
                  <a-doption value="edit"><icon-edit /> 编辑</a-doption>
                  <a-doption value="del"><icon-delete /> 删除</a-doption>
                </template>
              </a-dropdown>
            </div>
            <div v-if="savedConns.length === 0" class="empty-tip">暂无连接</div>
          </div>
        </div>

        <div class="lp-block tree-block" v-if="st.activeConnId">
          <div class="lp-head">
            <span class="lp-title">数据库</span>
            <a-button size="mini" type="text" @click="refreshDatabases()"><template
                #icon><icon-refresh /></template></a-button>
          </div>
          <div class="tree-wrap">
            <a-tree :data="treeData" v-model:selectedKeys="st.selectedKeys" :load-more="loadMore" blockNode
              @select="onTreeSelect" />
            <div v-if="treeData.length === 0" class="empty-tip">无数据库</div>
          </div>
        </div>
      </div>

      <!-- 右侧：主内容 -->
      <div class="main-panel">
        <div v-if="!st.activeConnId" class="empty-state">
          <icon-empty :size="56" class="empty-icon" />
          <div class="empty-text">请从左侧打开一个 MySQL 连接</div>
          <a-button type="outline" @click="openAddModal()"><template #icon><icon-plus /></template>新建连接</a-button>
        </div>

        <template v-else>
          <div class="crumb">
            <icon-storage /> {{ activeConnName }}
            <template v-if="st.activeDb"> / {{ st.activeDb }}</template>
            <template v-if="st.table"> / {{ st.table }}</template>
            <template v-if="st.view"> / {{ st.view }}</template>
            <template v-if="st.routine"> / {{ st.routine }}</template>
          </div>

          <!-- 右侧顶部工具栏：查询 / 备份 / 事务 -->
          <div class="main-toolbar">
            <a-button size="mini" :type="st.mode === 'query' ? 'primary' : 'outline'"
              @click="setMode('query')"><template #icon><icon-code /></template>查询</a-button>
            <a-button size="mini" :type="st.mode === 'backup' ? 'primary' : 'outline'"
              @click="setMode('backup')"><template #icon><icon-save /></template>备份</a-button>
            <a-button size="mini" :type="st.mode === 'tx' ? 'primary' : 'outline'" @click="setMode('tx')"><template
                #icon><icon-select /></template>事务</a-button>
          </div>

          <!-- 表：数据 -->
          <div v-if="st.mode === 'table'" class="content">
            <a-tabs v-model:activeKey="st.tableTab" size="small" type="rounded" @change="onTableTab">
              <a-tab-pane key="data" title="数据">
                <div class="grid-toolbar">
                  <a-input v-model="st.data.filter" size="mini" placeholder="WHERE 条件（不含 WHERE）" allow-clear
                    style="width: 280px" @press-enter="loadData()">
                    <template #prefix><icon-filter /></template>
                  </a-input>
                  <a-button size="mini" @click="loadData()"><template #icon><icon-search /></template>筛选</a-button>
                  <a-button size="mini" @click="clearFilter()"><template #icon><icon-refresh /></template>重置</a-button>
                  <a-button size="mini" type="primary" @click="openInsert()"><template
                      #icon><icon-plus /></template>新增行</a-button>
                  <a-button size="mini" @click="loadData()"><template #icon><icon-download /></template>刷新</a-button>
                  <span class="grid-note">{{ st.data.note }}</span>
                  <span v-if="!st.schema.pk" class="grid-warn">无单列主键，行编辑/删除已禁用</span>
                </div>
                <div class="grid-wrap">
                  <table class="data-grid">
                    <thead>
                      <tr>
                        <th class="rowno">#</th>
                        <th v-for="c in st.data.columns" :key="c" @click="sortBy(c)">
                          {{ c }}<span v-if="st.data.orderBy === c" class="sort-ind">{{ st.data.desc ? ' ↓' : ' ↑'
                            }}</span>
                        </th>
                        <th class="op"></th>
                      </tr>
                    </thead>
                    <tbody>
                      <tr v-for="(r, ri) in st.data.rows" :key="ri">
                        <td class="rowno">{{ (st.data.page - 1) * st.data.size + Number(ri) + 1 }}</td>
                        <td v-for="c in st.data.columns" :key="c" class="cell"
                          :class="{ editing: st.edit && st.edit.r === ri && st.edit.c === c }"
                          @dblclick="startEdit(Number(ri), c, r)">
                          <input v-if="st.edit && st.edit.r === ri && st.edit.c === c" v-model="st.edit.val" autofocus
                            @keyup.enter="commitEdit(r, c)" @keyup.esc="st.edit = null" @blur="commitEdit(r, c)" />
                          <span v-else :class="{ nullv: r[c] === null }">{{ r[c] === null ? 'NULL' : r[c] }}</span>
                        </td>
                        <td class="op">
                          <a-button size="mini" type="text" status="danger" :disabled="!st.schema.pk"
                            @click="deleteRow(r)"><icon-delete /></a-button>
                        </td>
                      </tr>
                      <tr v-if="st.data.rows.length === 0">
                        <td :colspan="st.data.columns.length + 2" class="empty-tip">无数据</td>
                      </tr>
                    </tbody>
                  </table>
                </div>
                <div class="pager">
                  <a-button size="mini" :disabled="st.data.page <= 1" @click="goPage(st.data.page - 1)"><template
                      #icon><icon-left /></template></a-button>
                  <span class="pager-info">第 {{ st.data.page }} / {{ totalPage }} 页（每页
                    <a-select v-model="st.data.size" size="mini" :options="sizeOptions" style="width: 78px"
                      @change="() => { st.data.page = 1; loadData(); }" />
                    ）</span>
                  <a-button size="mini" :disabled="st.data.page >= totalPage"
                    @click="goPage(st.data.page + 1)"><template #icon><icon-right /></template></a-button>
                </div>
              </a-tab-pane>

              <a-tab-pane key="struct" title="设计表">
                <div class="grid-toolbar">
                  <a-button size="mini" type="primary" @click="addColumnModal()"><template
                      #icon><icon-plus /></template>增加列</a-button>
                  <a-button size="mini" @click="createIndexModal()"><template
                      #icon><icon-plus /></template>新建索引</a-button>
                  <a-button size="mini" @click="loadSchema()"><template #icon><icon-refresh /></template>刷新</a-button>
                  <a-button size="mini" type="primary" status="danger" @click="dropTable()"><template
                      #icon><icon-delete /></template>删除表</a-button>
                </div>
                <div class="struct-title">列</div>
                <table class="info-grid">
                  <thead>
                    <tr>
                      <th>#</th>
                      <th>列名</th>
                      <th>类型</th>
                      <th>空</th>
                      <th>键</th>
                      <th>默认</th>
                      <th>额外</th>
                      <th>注释</th>
                      <th>操作</th>
                    </tr>
                  </thead>
                  <tbody>
                    <tr v-for="col in st.schema.columns" :key="col.name">
                      <td>{{ col.position }}</td>
                      <td>{{ col.name }}</td>
                      <td>{{ col.type }}</td>
                      <td>{{ col.nullable ? '是' : '否' }}</td>
                      <td>{{ col.key }}</td>
                      <td>{{ col.default }}</td>
                      <td>{{ col.extra }}</td>
                      <td>{{ col.comment }}</td>
                      <td class="op-cell">
                        <a-button size="mini" type="text" @click="editColumnModal(col)"><icon-edit /></a-button>
                        <a-button size="mini" type="text" status="danger"
                          @click="dropColumn(col)"><icon-delete /></a-button>
                      </td>
                    </tr>
                  </tbody>
                </table>
                <div class="struct-title">索引</div>
                <table class="info-grid">
                  <thead>
                    <tr>
                      <th>名称</th>
                      <th>唯一</th>
                      <th>主键</th>
                      <th>类型</th>
                      <th>列</th>
                      <th>操作</th>
                    </tr>
                  </thead>
                  <tbody>
                    <tr v-for="idx in st.schema.indexes" :key="idx.name">
                      <td>{{ idx.name }}</td>
                      <td>{{ idx.unique ? '是' : '否' }}</td>
                      <td>{{ idx.primary ? '是' : '否' }}</td>
                      <td>{{ idx.type }}</td>
                      <td>{{ (idx.columns || []).join(', ') }}</td>
                      <td><a-button size="mini" type="text" status="danger" :disabled="idx.primary"
                          @click="dropIndex(idx)">删除</a-button></td>
                    </tr>
                  </tbody>
                </table>
              </a-tab-pane>

              <a-tab-pane key="ddl" title="DDL">
                <div class="grid-toolbar"><a-button size="mini" @click="copyText(st.schema.ddl)"><template
                      #icon><icon-copy /></template>复制</a-button></div>
                <pre class="ddl-box">{{ st.schema.ddl }}</pre>
              </a-tab-pane>
            </a-tabs>
          </div>

          <!-- 视图 -->
          <div v-else-if="st.mode === 'view'" class="content">
            <div class="grid-toolbar">
              <span class="obj-title">视图：{{ st.view }}</span>
              <a-button size="mini" @click="loadViewData()"><template #icon><icon-refresh /></template>刷新数据</a-button>
              <a-button size="mini" status="danger" @click="dropView()"><template
                  #icon><icon-delete /></template>删除视图</a-button>
            </div>
            <SimpleGrid :columns="st.data.columns" :rows="st.data.rows" />
          </div>

          <!-- 函数 -->
          <div v-else-if="st.mode === 'routine'" class="content">
            <div class="grid-toolbar">
              <span class="obj-title">函数：{{ st.routine }}（{{ st.routineType }}）</span>
              <a-button size="mini" status="danger" @click="dropRoutine()"><template
                  #icon><icon-delete /></template>删除</a-button>
            </div>
            <div class="struct-title">定义</div>
            <pre class="ddl-box">{{ st.routineDef || '（无定义，可能是内置函数）' }}</pre>
          </div>

          <!-- 查询 -->
          <div v-else-if="st.mode === 'query'" class="content">
            <div class="grid-toolbar">
              <icon-thunderbolt class="tool-ic" />
              <a-select v-model="st.query.connId" size="mini" :options="openedConnOptions" style="width: 170px"
                placeholder="连接" @change="onPanelConnChange('query')" />
              <a-select v-model="st.query.db" size="mini" :options="queryDbOptions" style="width: 170px"
                placeholder="数据库" allow-search @change="onPanelDbChange('query')" />
              <a-button size="mini" type="primary" @click="runQuery()"><template #icon><icon-play-arrow /></template>运行
                (Ctrl+Enter)</a-button>
              <span class="grid-note">{{ st.query.note }}</span>
            </div>
            <a-textarea v-model="st.query.sql" :auto-size="{ minRows: 5, maxRows: 12 }"
              placeholder="输入 SQL 语句，例如 SELECT * FROM xxx LIMIT 100;" @keydown="onQueryKey" />
            <SimpleGrid :columns="st.query.result.columns" :rows="st.query.result.rows" />
          </div>

          <!-- 备份 -->
          <div v-else-if="st.mode === 'backup'" class="content">
            <div class="grid-toolbar">
              <icon-save class="tool-ic" />
              <a-select v-model="st.backup.connId" size="mini" :options="openedConnOptions" style="width: 170px"
                placeholder="连接" @change="onPanelConnChange('backup')" />
              <a-select v-model="st.backup.db" size="mini" :options="backupDbOptions" style="width: 170px"
                placeholder="数据库" allow-search @change="onPanelDbChange('backup')" />
              <a-checkbox v-model="st.backup.all">全部表</a-checkbox>
              <a-select v-if="!st.backup.all" v-model="st.backup.tables" :options="backupTableOptions" multiple
                placeholder="选择表" size="mini" style="width: 300px" allow-clear />
              <a-button size="mini" type="primary" :loading="st.backup.loading" @click="genBackup()"><template
                  #icon><icon-save /></template>生成备份</a-button>
              <a-button size="mini" :disabled="!st.backup.sql" @click="downloadBackup()"><template
                  #icon><icon-download /></template>下载 .sql</a-button>
            </div>
            <div class="struct-title">备份预览（{{ st.backup.sql.length }} 字符）</div>
            <pre class="ddl-box">{{ st.backup.sql || '点击「生成备份」导出当前数据库的 SQL 转储。' }}</pre>
          </div>

          <!-- 事务 -->
          <div v-else-if="st.mode === 'tx'" class="content">
            <div class="grid-toolbar">
              <icon-select class="tool-ic" />
              <a-select v-model="st.tx.connId" size="mini" :options="openedConnOptions" style="width: 170px"
                placeholder="连接" :disabled="!!st.tx.id" @change="onPanelConnChange('tx')" />
              <a-select v-model="st.tx.db" size="mini" :options="txDbOptions" style="width: 170px" placeholder="数据库"
                :disabled="!!st.tx.id" allow-search @change="onPanelDbChange('tx')" />
              <a-tag v-if="st.tx.id" color="green">事务进行中 {{ st.tx.id }}</a-tag>
              <a-button size="mini" type="primary" :disabled="!!st.tx.id" @click="txBegin()"><template
                  #icon><icon-play-arrow /></template>开始事务</a-button>
              <a-button size="mini" status="success" :disabled="!st.tx.id" @click="txCommit()"><template
                  #icon><icon-check /></template>提交</a-button>
              <a-button size="mini" status="warning" :disabled="!st.tx.id" @click="txRollback()"><template
                  #icon><icon-close /></template>回滚</a-button>
              <a-button size="mini" :disabled="!st.tx.id" @click="txRun()"><template
                  #icon><icon-code /></template>执行语句</a-button>
            </div>
            <a-textarea v-model="st.tx.sql" :auto-size="{ minRows: 4, maxRows: 10 }" :disabled="!st.tx.id"
              placeholder="在事务中执行 SQL（可多次），最后提交或回滚" />
            <SimpleGrid :columns="st.tx.result.columns" :rows="st.tx.result.rows" />
            <div class="struct-title">事务日志</div>
            <pre class="ddl-box tx-log">{{ st.tx.log.join('\n') || '（暂无）' }}</pre>
          </div>

          <div v-else class="empty-state">
            <icon-apps :size="48" class="empty-icon" />
            <div class="empty-text">选择左侧的表/视图/函数，或使用上方「查询 / 备份 / 事务」</div>
          </div>
        </template>
      </div>
    </div>

    <!-- 连接编辑 -->
    <a-modal v-model:visible="st.connModal" :title="st.connEditing ? '编辑连接' : '新建连接'" :width="480" @ok="saveConnection">
      <a-form :model="connForm" layout="vertical">
        <a-form-item field="name" label="连接名称"><a-input v-model="connForm.name"
            placeholder="例如 localhost_3306" /></a-form-item>
        <a-row :gutter="12">
          <a-col :span="16"><a-form-item field="host" label="主机"><a-input v-model="connForm.host"
                placeholder="127.0.0.1" /></a-form-item></a-col>
          <a-col :span="8"><a-form-item field="port" label="端口"><a-input-number v-model="connForm.port" :min="1"
                :max="65535" /></a-form-item></a-col>
        </a-row>
        <a-row :gutter="12">
          <a-col :span="12"><a-form-item field="user" label="用户名"><a-input v-model="connForm.user"
                placeholder="root" /></a-form-item></a-col>
          <a-col :span="12"><a-form-item field="password" label="密码"><a-input-password v-model="connForm.password"
                placeholder="密码" /></a-form-item></a-col>
        </a-row>
        <a-form-item field="database" label="默认数据库（可留空）"><a-input v-model="connForm.database"
            placeholder="留空则在左侧选择" /></a-form-item>
        <a-form-item label="保存位置">
          <a-radio-group v-model="st.saveTarget" type="button" size="small">
            <a-radio value="file"><icon-save /> 文件</a-radio>
            <a-radio value="cache"><icon-storage /> 缓存</a-radio>
          </a-radio-group>
        </a-form-item>
      </a-form>
    </a-modal>

    <!-- 新增行 -->
    <a-modal v-model:visible="st.insertModal" title="新增行" :width="520" @ok="doInsert">
      <a-form layout="vertical" :model="insertData">
        <a-form-item v-for="col in insertCols" :key="col.name" :label="col.name + ' (' + col.type + ')'">
          <div style="display:flex; gap:8px; align-items:center">
            <a-input v-model="insertData[col.name]" :disabled="insertNull[col.name]"
              :placeholder="col.comment || col.extra || ''" />
            <a-checkbox v-model="insertNull[col.name]">NULL</a-checkbox>
          </div>
        </a-form-item>
      </a-form>
    </a-modal>

    <!-- 新建索引 -->
    <a-modal v-model:visible="st.indexModal" title="新建索引" :width="460" @ok="doCreateIndex">
      <a-form layout="vertical" :model="indexForm">
        <a-form-item label="索引名称"><a-input v-model="indexForm.name" placeholder="idx_xxx" /></a-form-item>
        <a-form-item label="列（逗号分隔）"><a-input v-model="indexForm.columns" placeholder="col1, col2" /></a-form-item>
        <a-form-item><a-checkbox v-model="indexForm.unique">唯一索引</a-checkbox></a-form-item>
      </a-form>
    </a-modal>

    <!-- 新增/编辑列 -->
    <a-modal v-model:visible="st.colModal" :title="st.colEditing ? '编辑列：' + st.colOldName : '增加列'" :width="500"
      @ok="saveColumn">
      <a-form layout="vertical" :model="colForm">
        <a-row :gutter="12">
          <a-col :span="12"><a-form-item label="列名"><a-input v-model="colForm.name"
                placeholder="column_name" /></a-form-item></a-col>
          <a-col :span="12"><a-form-item label="类型"><a-input v-model="colForm.type"
                placeholder="如 varchar(255) / int / datetime" /></a-form-item></a-col>
        </a-row>
        <a-form-item label="默认值（留空=无；可填 NULL / CURRENT_TIMESTAMP）"><a-input v-model="colForm.default" /></a-form-item>
        <a-form-item label="注释"><a-input v-model="colForm.comment" /></a-form-item>
        <a-form-item label="位置（置于某列之后，默认末尾）"><a-select v-model="colForm.after" :options="afterOptions" allow-clear
            placeholder="默认末尾" /></a-form-item>
        <a-space :size="20">
          <a-checkbox v-model="colForm.nullable">允许空 (NULL)</a-checkbox>
          <a-checkbox v-model="colForm.autoIncrement">自增</a-checkbox>
          <a-checkbox v-model="colForm.primaryKey">主键</a-checkbox>
        </a-space>
      </a-form>
    </a-modal>

    <!-- 新建表 -->
    <a-modal v-model:visible="st.newTableModal" title="新建表 (DDL)" :width="560" @ok="doCreateTable">
      <a-textarea v-model="st.newTableDDL" :auto-size="{ minRows: 8 }"
        placeholder="CREATE TABLE my_table (&#10;  id INT AUTO_INCREMENT PRIMARY KEY,&#10;  name VARCHAR(64),&#10;  created DATETIME&#10;) ENGINE=InnoDB;" />
    </a-modal>
  </layout-body>
</template>

<script setup lang="ts">
import { computed, getCurrentInstance, h, onMounted, reactive, ref } from "vue";
import { Message, Modal } from "@arco-design/web-vue";

const GoTools = (getCurrentInstance() as any)?.appContext.config.globalProperties.$GoTools;
const CACHE_KEY = "mysql_connections_cache";

/* 简单只读结果表格（视图/查询/事务复用） */
const SimpleGrid = (props: { columns: string[]; rows: any[] }) =>
  h("div", { class: "grid-wrap" }, [
    h("table", { class: "data-grid" }, [
      h("thead", [h("tr", [h("th", { class: "rowno" }, "#"), ...(props.columns || []).map((c) => h("th", c))])]),
      h("tbody", (props.rows || []).map((r, i) =>
        h("tr", [h("td", { class: "rowno" }, String(i + 1)), ...(props.columns || []).map((c) => h("td", { class: "cell" }, r[c] === null ? "NULL" : String(r[c])))]),
      )),
    ]),
  ]);
SimpleGrid.props = ["columns", "rows"];

interface ConnConfig { id: string; name: string; host: string; port: number; user: string; password: string; database: string }

const savedConns = ref<ConnConfig[]>([]);
const openedIds = ref<string[]>([]);

const emptyResult = () => ({ columns: [] as string[], rows: [] as any[], affected: 0, lastID: 0, elapsed: 0, note: "" });

const st = reactive({
  activeConnId: "",
  activeDb: "",
  databases: [] as string[],
  tables: [] as any[],
  views: [] as any[],
  routines: [] as any[],
  selectedKeys: [] as string[],
  mode: "none", // none|table|view|routine|query|backup|tx
  tableTab: "data",
  table: "",
  view: "",
  routine: "",
  routineType: "",
  routineDef: "",
  schema: { columns: [], indexes: [], ddl: "", pk: "" } as any,
  data: { columns: [], rows: [], total: 0, page: 1, size: 50, filter: "", orderBy: "", desc: false, note: "" } as any,
  edit: null as null | { r: number; c: string; val: string },
  query: { connId: "", db: "", databases: [] as string[], sql: "", result: emptyResult(), note: "" } as any,
  tx: { connId: "", db: "", databases: [] as string[], id: "", sql: "", result: emptyResult(), log: [] as string[] } as any,
  backup: { connId: "", db: "", databases: [] as string[], tableList: [] as any[], all: true, tables: [] as string[], sql: "", loading: false } as any,
  // 弹窗
  connModal: false,
  connEditing: false,
  saveTarget: "file",
  insertModal: false,
  indexModal: false,
  newTableModal: false,
  newTableDDL: "",
  colModal: false,
  colEditing: false,
  colOldName: "",
});

const connForm = reactive<ConnConfig>({ id: "", name: "", host: "127.0.0.1", port: 3306, user: "root", password: "", database: "" });
const insertData = reactive<Record<string, string>>({});
const insertNull = reactive<Record<string, boolean>>({});
const indexForm = reactive({ name: "", columns: "", unique: false });
const colForm = reactive({ name: "", type: "varchar(255)", nullable: true, default: "", autoIncrement: false, primaryKey: false, comment: "", after: "" });

const activeConnName = computed(() => savedConns.value.find((c) => c.id === st.activeConnId)?.name || "");
const isOpen = (id: string) => openedIds.value.includes(id);
const openedConnOptions = computed(() =>
  savedConns.value.filter((c) => openedIds.value.includes(c.id)).map((c) => ({ label: c.name || `${c.host}:${c.port}`, value: c.id })));
const queryDbOptions = computed(() => (st.query.databases || []).map((d: string) => ({ label: d, value: d })));
const backupDbOptions = computed(() => (st.backup.databases || []).map((d: string) => ({ label: d, value: d })));
const txDbOptions = computed(() => (st.tx.databases || []).map((d: string) => ({ label: d, value: d })));
const backupTableOptions = computed(() => (st.backup.tableList || []).map((t: any) => ({ label: t.name, value: t.name })));
const afterOptions = computed(() => (st.schema.columns || []).map((c: any) => ({ label: c.name, value: c.name })));
const insertCols = computed<any[]>(() => (st.schema.columns || []).filter((c: any) => !(c.extra || "").toLowerCase().includes("auto_increment")));
const totalPage = computed(() => Math.max(1, Math.ceil((st.data.total || 0) / (st.data.size || 50))));
const sizeOptions = [25, 50, 100, 200, 500].map((n) => ({ label: String(n), value: n }));

const treeData = ref<any[]>([]);
/** 树节点 key → 元信息，避免解析标识符（库/表名可能含特殊字符） */
const nodeMap = new Map<string, { kind: "db" | "table" | "view" | "func"; db: string; name: string }>();

function buildGroups(db: string, t: any[], v: any[], r: any[]) {
  const groups: any[] = [];
  if (t.length) groups.push({ key: `grp:${db}:t`, title: `表 (${t.length})`, isLeaf: false, children: t.map((x) => { const k = `table:${db}:${x.name}`; nodeMap.set(k, { kind: "table", db, name: x.name }); return { key: k, title: x.name, isLeaf: true }; }) });
  if (v.length) groups.push({ key: `grp:${db}:v`, title: `视图 (${v.length})`, isLeaf: false, children: v.map((x) => { const k = `view:${db}:${x.name}`; nodeMap.set(k, { kind: "view", db, name: x.name }); return { key: k, title: x.name, isLeaf: true }; }) });
  if (r.length) groups.push({ key: `grp:${db}:r`, title: `函数 (${r.length})`, isLeaf: false, children: r.map((x) => { const k = `func:${db}:${x.name}`; nodeMap.set(k, { kind: "func", db, name: x.name }); return { key: k, title: x.name, isLeaf: true }; }) });
  if (groups.length === 0) groups.push({ key: `grp:${db}:empty`, title: "（空）", isLeaf: true, selectable: false });
  return groups;
}

/* ============================ 连接 ============================ */

async function loadConnections() {
  let list: ConnConfig[] = [];
  try { list = (await GoTools.MySQLLoadConnections()) || []; } catch (e) { list = []; }
  if (!list || list.length === 0) {
    try { const c = localStorage.getItem(CACHE_KEY); if (c) list = JSON.parse(c) || []; } catch (e) { list = []; }
  }
  savedConns.value = (list || []).map((x) => ({ ...x }));
}

async function persistConnections() {
  const data = JSON.parse(JSON.stringify(savedConns.value));
  localStorage.setItem(CACHE_KEY, JSON.stringify(data));
  if (st.saveTarget === "file") {
    try { await GoTools.MySQLSaveConnections(data); } catch (e) { /* ignore */ }
  }
}

function openAddModal() {
  st.connEditing = false;
  Object.assign(connForm, { id: "", name: "", host: "127.0.0.1", port: 3306, user: "root", password: "", database: "" });
  st.connModal = true;
}
function openEditModal(c: ConnConfig) {
  st.connEditing = true;
  Object.assign(connForm, { ...c });
  st.connModal = true;
}
async function saveConnection() {
  if (!connForm.host) { Message.warning("请填写主机"); return; }
  const item: ConnConfig = { ...connForm };
  if (!item.name) item.name = `${item.host}:${item.port}`;
  if (st.connEditing) {
    const i = savedConns.value.findIndex((c) => c.id === item.id);
    if (i >= 0) savedConns.value[i] = item;
  } else {
    item.id = "m" + Date.now().toString(36);
    savedConns.value.push(item);
  }
  await persistConnections();
  st.connModal = false;
  Message.success("已保存到" + (st.saveTarget === "file" ? "文件" : "缓存"));
}
function onConnAction(v: string, c: ConnConfig) {
  if (v === "open") openConnection(c);
  else if (v === "edit") openEditModal(c);
  else if (v === "del") confirmDeleteConn(c);
}
function confirmDeleteConn(c: ConnConfig) {
  Modal.warning({
    title: "删除连接", content: `确定删除「${c.name}」？`, hideCancel: false, onOk: async () => {
      if (isOpen(c.id)) { GoTools.MySQLClose(c.id); openedIds.value = openedIds.value.filter((x) => x !== c.id); }
      if (st.activeConnId === c.id) resetActive();
      savedConns.value = savedConns.value.filter((x) => x.id !== c.id);
      await persistConnections();
    }
  });
}

async function openConnection(c: ConnConfig) {
  try {
    await GoTools.MySQLOpen({ ...c });
  } catch (e: any) { Message.error(String(e?.message || e)); return; }
  if (!isOpen(c.id)) openedIds.value.push(c.id);
  st.activeConnId = c.id;
  st.mode = "none";
  if (c.database) {
    st.activeDb = c.database;
  } else {
    st.activeDb = "";
  }
  await refreshDatabases();
  if (st.activeDb) { await reloadDb(st.activeDb); st.selectedKeys = ["db:" + st.activeDb]; }
}

function resetActive() {
  st.activeConnId = ""; st.activeDb = ""; st.tables = []; st.views = []; st.routines = [];
  st.databases = []; st.mode = "none"; st.table = "";
}

async function refreshDatabases() {
  try { st.databases = (await GoTools.MySQLListDatabases(st.activeConnId)) || []; }
  catch (e: any) { Message.error(String(e?.message || e)); return; }
  nodeMap.clear();
  treeData.value = st.databases.map((db) => {
    const key = "db:" + db;
    nodeMap.set(key, { kind: "db", db, name: db });
    return { key, title: db, isLeaf: false };
  });
}

/** 拉取指定库的表/视图/函数 */
async function fetchObjects(db: string) {
  const [t, v, r] = await Promise.all([
    GoTools.MySQLListTables(st.activeConnId, db),
    GoTools.MySQLListViews(st.activeConnId, db),
    GoTools.MySQLListRoutines(st.activeConnId, db),
  ]);
  return { t: t || [], v: v || [], r: r || [] };
}

/** 展开数据库节点时懒加载其表/视图/函数 */
function loadMore(node: any) {
  const meta = nodeMap.get(node.key);
  if (!meta || meta.kind !== "db") return Promise.resolve();
  return fetchObjects(meta.db)
    .then(({ t, v, r }) => {
      node.children = buildGroups(meta.db, t, v, r);
      if (meta.db === st.activeDb) { st.tables = t; st.views = v; st.routines = r; }
    })
    .catch((e: any) => { Message.error(String(e?.message || e)); });
}

/** 重新加载当前库对象，并同步树节点（若已展开） */
async function reloadDb(db: string) {
  try {
    const { t, v, r } = await fetchObjects(db);
    st.tables = t; st.views = v; st.routines = r;
    const node = treeData.value.find((n: any) => n.key === "db:" + db);
    if (node && node.children) node.children = buildGroups(db, t, v, r);
  } catch (e: any) { Message.error(String(e?.message || e)); }
}

function loadObjects() { return reloadDb(st.activeDb); }

async function setMode(m: string) {
  st.mode = m;
  if (m === "query" || m === "backup" || m === "tx") {
    const p: any = (st as any)[m];
    p.connId = st.activeConnId;
    p.db = st.activeDb;
    await panelLoadDatabases(m);
    if (m === "backup") await panelLoadTables(m);
  }
}

/** 加载某个面板（查询/备份/事务）所选连接的数据库列表 */
async function panelLoadDatabases(m: string) {
  const p: any = (st as any)[m];
  if (!p.connId) { p.databases = []; return; }
  try { p.databases = (await GoTools.MySQLListDatabases(p.connId)) || []; }
  catch (e: any) { p.databases = []; Message.error(String(e?.message || e)); }
}

/** 备份面板：加载所选库的表列表 */
async function panelLoadTables(m: string) {
  const p: any = (st as any)[m];
  if (m !== "backup") return;
  if (!p.connId || !p.db) { p.tableList = []; return; }
  try { p.tableList = (await GoTools.MySQLListTables(p.connId, p.db)) || []; }
  catch { p.tableList = []; }
}

async function onPanelConnChange(m: string) {
  const p: any = (st as any)[m];
  p.db = "";
  if (m === "backup") { p.tables = []; p.tableList = []; p.sql = ""; }
  await panelLoadDatabases(m);
}

async function onPanelDbChange(m: string) {
  if (m === "backup") { st.backup.tables = []; await panelLoadTables("backup"); }
}

function onTreeSelect(keys: (string | number)[]) {
  if (!keys || keys.length === 0) return;
  const meta = nodeMap.get(String(keys[0]));
  if (!meta) return;
  if (meta.kind === "db") { st.activeDb = meta.db; st.mode = "none"; st.table = ""; reloadDb(meta.db); }
  else if (meta.kind === "table") { st.activeDb = meta.db; openTable(meta.name); }
  else if (meta.kind === "view") { st.activeDb = meta.db; openView(meta.name); }
  else if (meta.kind === "func") { st.activeDb = meta.db; openRoutine(meta.name); }
}

/* ============================ 表 ============================ */

async function openTable(name: string) {
  st.table = name; st.mode = "table"; st.tableTab = "data";
  st.data.filter = ""; st.data.orderBy = ""; st.data.desc = false; st.data.page = 1;
  st.schema = { columns: [], indexes: [], ddl: "", pk: "" };
  await loadSchema();
  await loadData();
}

async function loadSchema() {
  try { st.schema = await GoTools.MySQLTableSchema(st.activeConnId, st.activeDb, st.table); }
  catch (e: any) { Message.error(String(e?.message || e)); }
}

async function loadData() {
  try {
    const res = await GoTools.MySQLTableData(st.activeConnId, st.activeDb, st.table, st.data.filter, st.data.orderBy, st.data.desc, st.data.page, st.data.size);
    st.data.columns = res.columns || [];
    st.data.rows = res.rows || [];
    st.data.total = res.affected || 0;
    st.data.note = res.note || "";
    st.edit = null;
  } catch (e: any) { Message.error(String(e?.message || e)); }
}

function onTableTab(k: string | number) { if (String(k) === "data" && st.data.rows.length === 0) loadData(); }
function clearFilter() { st.data.filter = ""; st.data.orderBy = ""; st.data.desc = false; st.data.page = 1; loadData(); }
function goPage(p: number) { st.data.page = p; loadData(); }
function sortBy(c: string) { if (st.data.orderBy === c) st.data.desc = !st.data.desc; else { st.data.orderBy = c; st.data.desc = false; } st.data.page = 1; loadData(); }

function startEdit(ri: number, c: string, row: any) {
  if (!st.schema.pk) { Message.warning("无单列主键，无法编辑"); return; }
  st.edit = { r: ri, c, val: row[c] === null ? "" : String(row[c]) };
}
async function commitEdit(row: any, c: string) {
  if (!st.edit) return;
  const val = st.edit.val;
  st.edit = null;
  const cur = row[c] === null ? "" : String(row[c]);
  if (val === cur) return;
  try {
    await GoTools.MySQLUpdateCell(st.activeConnId, st.activeDb, st.table, st.schema.pk, String(row[st.schema.pk]), c, val);
    Message.success("已更新");
    loadData();
  } catch (e: any) { Message.error(String(e?.message || e)); }
}

function openInsert() {
  if (!st.schema.columns.length) { Message.warning("请先加载表结构"); return; }
  Object.keys(insertData).forEach((k) => delete insertData[k]);
  Object.keys(insertNull).forEach((k) => delete insertNull[k]);
  insertCols.value.forEach((c: any) => { insertData[c.name] = ""; insertNull[c.name] = false; });
  st.insertModal = true;
}
async function doInsert() {
  const data: Record<string, string> = {};
  insertCols.value.forEach((c: any) => {
    if (insertNull[c.name]) data[c.name] = "\u0000NULL";
    else if (insertData[c.name] !== "") data[c.name] = insertData[c.name];
  });
  if (Object.keys(data).length === 0) { Message.warning("请填写至少一个字段"); return; }
  try {
    await GoTools.MySQLInsertRow(st.activeConnId, st.activeDb, st.table, data);
    Message.success("已新增");
    st.insertModal = false;
    loadData();
  } catch (e: any) { Message.error(String(e?.message || e)); }
}

function deleteRow(row: any) {
  if (!st.schema.pk) return;
  Modal.warning({
    title: "删除行", content: `确定删除主键 ${st.schema.pk}=${row[st.schema.pk]} 的行？`, hideCancel: false, onOk: async () => {
      try { await GoTools.MySQLDeleteRow(st.activeConnId, st.activeDb, st.table, st.schema.pk, String(row[st.schema.pk])); loadData(); }
      catch (e: any) { Message.error(String(e?.message || e)); }
    }
  });
}

function dropTable() {
  Modal.warning({
    title: "删除表", content: `确定删除表「${st.table}」？此操作不可恢复。`, hideCancel: false, onOk: async () => {
      try { await GoTools.MySQLExecute(st.activeConnId, st.activeDb, "DROP TABLE " + bt(st.table)); st.mode = "none"; st.table = ""; await loadObjects(); }
      catch (e: any) { Message.error(String(e?.message || e)); }
    }
  });
}

function bt(s: string) { return "`" + String(s).replace(/`/g, "``") + "`"; }

function createIndexModal() { Object.assign(indexForm, { name: "", columns: "", unique: false }); st.indexModal = true; }
async function doCreateIndex() {
  if (!indexForm.name || !indexForm.columns) { Message.warning("请填写索引名和列"); return; }
  const cols = indexForm.columns.split(",").map((s) => bt(s.trim())).join(",");
  const sql = `CREATE ${indexForm.unique ? "UNIQUE " : ""}INDEX ${bt(indexForm.name)} ON ${bt(st.table)} (${cols})`;
  try { await GoTools.MySQLExecute(st.activeConnId, st.activeDb, sql); Message.success("索引已创建"); st.indexModal = false; loadSchema(); }
  catch (e: any) { Message.error(String(e?.message || e)); }
}
function dropIndex(idx: any) {
  Modal.warning({
    title: "删除索引", content: `确定删除索引「${idx.name}」？`, hideCancel: false, onOk: async () => {
      try { await GoTools.MySQLExecute(st.activeConnId, st.activeDb, `DROP INDEX ${bt(idx.name)} ON ${bt(st.table)}`); loadSchema(); }
      catch (e: any) { Message.error(String(e?.message || e)); }
    }
  });
}

/* ---- 列增/改/删 ---- */
function addColumnModal() {
  st.colEditing = false; st.colOldName = "";
  Object.assign(colForm, { name: "", type: "varchar(255)", nullable: true, default: "", autoIncrement: false, primaryKey: false, comment: "", after: "" });
  st.colModal = true;
}
function editColumnModal(col: any) {
  st.colEditing = true; st.colOldName = col.name;
  Object.assign(colForm, {
    name: col.name, type: col.type, nullable: !!col.nullable,
    default: col.default || "",
    autoIncrement: (col.extra || "").toLowerCase().includes("auto_increment"),
    primaryKey: col.key === "PRI",
    comment: col.comment || "", after: "",
  });
  st.colModal = true;
}
async function saveColumn() {
  if (!colForm.name || !colForm.type) { Message.warning("列名与类型必填"); return; }
  const payload = { ...colForm };
  try {
    if (st.colEditing) await GoTools.MySQLModifyColumn(st.activeConnId, st.activeDb, st.table, st.colOldName, payload);
    else await GoTools.MySQLAddColumn(st.activeConnId, st.activeDb, st.table, payload);
    Message.success("已保存"); st.colModal = false; await loadSchema();
  } catch (e: any) { Message.error(String(e?.message || e)); }
}
function dropColumn(col: any) {
  Modal.warning({
    title: "删除列", content: `确定删除列「${col.name}」？该操作不可恢复。`, hideCancel: false, onOk: async () => {
      try { await GoTools.MySQLDropColumn(st.activeConnId, st.activeDb, st.table, col.name); await loadSchema(); }
      catch (e: any) { Message.error(String(e?.message || e)); }
    }
  });
}

function doCreateTable() {
  if (!st.newTableDDL.trim()) { Message.warning("请输入 DDL"); return; }
  GoTools.MySQLExecute(st.activeConnId, st.activeDb, st.newTableDDL).then(() => {
    Message.success("表已创建"); st.newTableModal = false; loadObjects();
  }).catch((e: any) => Message.error(String(e?.message || e)));
}

/* ============================ 视图 / 函数 ============================ */

function openView(name: string) {
  st.view = name; st.mode = "view";
  loadViewData();
}
async function loadViewData() {
  try {
    const res = await GoTools.MySQLExecute(st.activeConnId, st.activeDb, "SELECT * FROM " + bt(st.view) + " LIMIT 500");
    st.data.columns = res.columns || []; st.data.rows = res.rows || [];
  } catch (e: any) { Message.error(String(e?.message || e)); }
}
function dropView() {
  Modal.warning({
    title: "删除视图", content: `确定删除视图「${st.view}」？`, hideCancel: false, onOk: async () => {
      try { await GoTools.MySQLExecute(st.activeConnId, st.activeDb, "DROP VIEW " + bt(st.view)); st.mode = "none"; st.view = ""; await loadObjects(); }
      catch (e: any) { Message.error(String(e?.message || e)); }
    }
  });
}

function openRoutine(name: string) {
  const r = st.routines.find((x) => x.name === name);
  st.routine = name; st.routineType = r?.type || ""; st.routineDef = r?.definition || ""; st.mode = "routine";
}
function dropRoutine() {
  Modal.warning({
    title: "删除", content: `确定删除${st.routineType === "PROCEDURE" ? "存储过程" : "函数"}「${st.routine}」？`, hideCancel: false, onOk: async () => {
      try {
        const kind = st.routineType === "PROCEDURE" ? "PROCEDURE" : "FUNCTION";
        await GoTools.MySQLExecute(st.activeConnId, st.activeDb, `DROP ${kind} ${bt(st.routine)}`);
        st.mode = "none"; st.routine = ""; await loadObjects();
      } catch (e: any) { Message.error(String(e?.message || e)); }
    }
  });
}

/* ============================ 查询 ============================ */

async function runQuery() {
  if (!st.query.connId) { Message.warning("请先选择连接"); return; }
  if (!st.query.sql.trim()) { Message.warning("请输入 SQL"); return; }
  st.query.note = "执行中…";
  try {
    const res = await GoTools.MySQLExecute(st.query.connId, st.query.db, st.query.sql);
    st.query.result = res;
    st.query.note = `${res.note} · ${res.elapsed}ms`;
  } catch (e: any) { st.query.note = ""; Message.error(String(e?.message || e)); }
}
function onQueryKey(e: KeyboardEvent) { if (e.ctrlKey && e.key === "Enter") { e.preventDefault(); runQuery(); } }

/* ============================ 备份 ============================ */

async function genBackup() {
  if (!st.backup.connId) { Message.warning("请先选择连接"); return; }
  if (!st.backup.db) { Message.warning("请先选择数据库"); return; }
  st.backup.loading = true;
  try {
    const tables = st.backup.all ? [] : st.backup.tables;
    st.backup.sql = await GoTools.MySQLBackup(st.backup.connId, st.backup.db, tables);
    Message.success("备份已生成");
  } catch (e: any) { Message.error(String(e?.message || e)); }
  st.backup.loading = false;
}
function downloadBackup() {
  const blob = new Blob([st.backup.sql], { type: "text/plain;charset=utf-8" });
  const url = URL.createObjectURL(blob);
  const a = document.createElement("a");
  a.href = url; a.download = `${st.backup.db}_backup_${Date.now()}.sql`;
  document.body.appendChild(a); a.click(); document.body.removeChild(a);
  URL.revokeObjectURL(url);
}

/* ============================ 事务 ============================ */

async function txBegin() {
  if (!st.tx.connId) { Message.warning("请先选择连接"); return; }
  try {
    st.tx.id = await GoTools.MySQLBeginTx(st.tx.connId, st.tx.db);
    st.tx.log.push("BEGIN -> " + st.tx.id);
    st.tx.result = emptyResult();
  } catch (e: any) { Message.error(String(e?.message || e)); }
}
async function txRun() {
  if (!st.tx.id || !st.tx.sql.trim()) return;
  try {
    const res = await GoTools.MySQLTxExecute(st.tx.id, st.tx.sql);
    st.tx.result = res;
    st.tx.log.push("EXEC " + st.tx.sql.split("\n")[0].slice(0, 60) + " -> " + res.note);
  } catch (e: any) { st.tx.log.push("ERROR " + (e?.message || e)); Message.error(String(e?.message || e)); }
}
async function txCommit() {
  try { await GoTools.MySQLCommitTx(st.tx.id); st.tx.log.push("COMMIT"); Message.success("已提交"); st.tx.id = ""; }
  catch (e: any) { Message.error(String(e?.message || e)); }
}
async function txRollback() {
  try { await GoTools.MySQLRollbackTx(st.tx.id); st.tx.log.push("ROLLBACK"); Message.success("已回滚"); st.tx.id = ""; }
  catch (e: any) { Message.error(String(e?.message || e)); }
}

/* ============================ 工具 ============================ */

function copyText(t: string) {
  const el = document.createElement("textarea"); el.value = t || ""; document.body.appendChild(el); el.select();
  document.execCommand("copy"); document.body.removeChild(el); Message.success("已复制");
}

onMounted(loadConnections);
</script>

<style scoped>
.mysql-page {
  display: flex;
  height: 100%;
  min-height: 0;
  background: var(--color-fill-1);
}

.left-panel {
  width: 240px;
  flex-shrink: 0;
  border-right: 1px solid var(--color-border-2);
  background: var(--color-bg-2);
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.lp-block {
  padding: 8px;
  border-bottom: 1px solid var(--color-border-2);
}

.tree-block {
  flex: 1;
  display: flex;
  flex-direction: column;
  min-height: 0;
  overflow: hidden;
}

.tree-wrap {
  flex: 1;
  overflow: auto;
}

.lp-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 4px;
}

.lp-title {
  font-weight: 600;
  font-size: 13px;
}

.conn-list {
  max-height: 160px;
  overflow-y: auto;
}

.conn-item {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 6px 8px;
  border-radius: 4px;
  cursor: pointer;
  font-size: 13px;
}

.conn-item:hover {
  background: var(--color-fill-3);
}

.conn-item.active {
  background: var(--color-primary-light-1);
}

.conn-name {
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.conn-more {
  color: var(--color-text-3);
}

.led {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: #c9cdd4;
  flex-shrink: 0;
}

.led.on {
  background: #00b42a;
  box-shadow: 0 0 4px #00b42a;
}

.quick-actions {
  display: flex;
  gap: 4px;
  margin-top: 8px;
}

.empty-tip {
  color: var(--color-text-3);
  text-align: center;
  font-size: 12px;
  padding: 12px;
}

.main-panel {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  background: var(--color-bg-1);
}

.empty-state {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 12px;
  color: var(--color-text-3);
}

.empty-icon {
  color: var(--color-text-4);
}

.crumb {
  padding: 8px 12px;
  border-bottom: 1px solid var(--color-border-2);
  font-size: 13px;
  color: var(--color-text-2);
  background: var(--color-bg-2);
}

.main-toolbar {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 12px;
  border-bottom: 1px solid var(--color-border-2);
  background: var(--color-bg-2);
}

.tool-ic {
  color: var(--color-text-2);
}

.content {
  flex: 1;
  overflow: auto;
  padding: 10px 12px;
}

.grid-toolbar {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 8px;
  flex-wrap: wrap;
}

.grid-note {
  color: var(--color-text-3);
  font-size: 12px;
}

.grid-warn {
  color: rgb(var(--orange-6));
  font-size: 12px;
}

.obj-title {
  font-weight: 600;
}

.grid-wrap {
  border: 1px solid var(--color-border-2);
  border-radius: 4px;
  overflow: auto;
  max-height: 520px;
}

.data-grid {
  border-collapse: collapse;
  width: max-content;
  min-width: 100%;
  font-size: 13px;
}

.data-grid th,
.data-grid td {
  border: 1px solid var(--color-border-2);
  padding: 4px 10px;
  white-space: nowrap;
  max-width: 320px;
  overflow: hidden;
  text-overflow: ellipsis;
}

.data-grid thead th {
  position: sticky;
  top: 0;
  background: var(--color-fill-2);
  cursor: pointer;
  user-select: none;
  z-index: 1;
}

.data-grid .rowno {
  width: 44px;
  text-align: center;
  color: var(--color-text-3);
  background: var(--color-fill-1);
}

.data-grid .cell {
  cursor: cell;
}

.data-grid .cell.editing {
  padding: 0;
}

.data-grid .cell input {
  width: 100%;
  border: 2px solid rgb(var(--blue-6));
  outline: none;
  padding: 3px 8px;
  font-size: 13px;
}

.data-grid .nullv {
  color: var(--color-text-4);
  font-style: italic;
}

.data-grid .op {
  width: 40px;
  text-align: center;
}

.sort-ind {
  color: rgb(var(--blue-6));
}

.pager {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-top: 8px;
}

.pager-info {
  font-size: 12px;
  color: var(--color-text-2);
  display: flex;
  align-items: center;
  gap: 4px;
}

.info-grid {
  border-collapse: collapse;
  width: 100%;
  font-size: 13px;
  margin-bottom: 12px;
}

.info-grid th,
.info-grid td {
  border: 1px solid var(--color-border-2);
  padding: 4px 10px;
  text-align: left;
}

.info-grid thead th {
  background: var(--color-fill-2);
}

.op-cell {
  white-space: nowrap;
  text-align: center;
}

.struct-title {
  font-weight: 600;
  margin: 10px 0 6px;
  font-size: 13px;
}

.ddl-box {
  background: var(--color-fill-1);
  border: 1px solid var(--color-border-2);
  border-radius: 4px;
  padding: 10px;
  font-family: Consolas, monospace;
  font-size: 13px;
  white-space: pre-wrap;
  word-break: break-all;
  max-height: 460px;
  overflow: auto;
}

.tx-log {
  max-height: 200px;
}
</style>
