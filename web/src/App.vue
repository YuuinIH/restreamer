<template>
  <div class="common-layout">
    <el-menu class="el-menu-demo" mode="horizontal">
      <el-menu-item index="1">星霜转播机</el-menu-item>
    </el-menu>
    <div style="display: flex">
      <el-menu
        default-active="2"
        class="el-menu-vertical-demo"
        :collapse="true"
      >
        <el-menu-item index="1" @click="opencreate = true">
          <el-icon><plus /></el-icon>
          <template #title>添加</template>
        </el-menu-item>
        <el-menu-item index="2">
          <el-icon><RefreshRight /></el-icon>
          <template #title>刷新</template>
        </el-menu-item>
      </el-menu>
      <div style="padding: 20px;width: 100%;">
      <el-collapse>
        <div v-show="opencreate">
          <p>创建</p>
          <el-form :model="create" label-width="120px">
            <el-form-item label="名称">
              <el-input v-model="create.name" />
            </el-form-item>
            <el-form-item label="画面来源">
              <el-input v-model="create.sourceurl" />
            </el-form-item>
            <el-form-item label="目标推流">
              <el-input v-model="create.streamurl" />
            </el-form-item>
            <el-form-item label="自动重启">
              <el-switch v-model="create.autorestart" />
            </el-form-item>
          </el-form>
          <el-button-group>
            <el-button
              type="primary"
              :icon="Edit"
              @click="
                Save('', create);
                opencreate = false;
              "
              >保存</el-button
            >
            <el-button type="primary" @click="opencreate = false"
              >关闭</el-button
            >
          </el-button-group>
        </div>
        <el-collapse-item
          v-for="(item, name) in streamer"
          :name="item.name"
          :title="`${name.toString()} ${getstatename(item.status)}`"
        >
          <el-form :model="item" label-width="120px">
            <el-form-item label="名称">
              <el-input v-model="item.name" />
            </el-form-item>
            <el-form-item label="画面来源">
              <el-input v-model="item.sourceurl" />
            </el-form-item>
            <el-form-item label="目标推流">
              <el-input v-model="item.streamurl" />
            </el-form-item>
            <el-form-item label="自动重启">
              <el-switch v-model="item.autorestart" />
            </el-form-item>
          </el-form>
          <el-button-group>
            <el-button
              type="primary"
              :icon="Edit"
              @click="Save(name.toString(), item)"
              >保存</el-button
            >
            <el-button
              type="primary"
              :icon="CaretRight"
              @click="start(name.toString())"
              >启动</el-button
            >
            <el-button
              type="primary"
              :icon="VideoPause"
              @click="stop(name.toString())"
              >停止</el-button
            >
            <el-button
              type="danger"
              :icon="Delete"
              @click="del(name.toString())"
              >删除</el-button
            >
          </el-button-group>
        </el-collapse-item>
      </el-collapse>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from "vue";
import {
  Plus,
  Edit,
  RefreshRight,
  CaretRight,
  VideoPause,
  Delete,
} from "@element-plus/icons-vue";
interface Streamer {
  name: string;
  streamurl: string;
  sourceurl: string;
  status: number;
  autorestart: boolean;
}
interface Streamerlist {
  [index: string]: Streamer;
}
enum state {
  RUNNING,
  WAITING,
  PAUSE,
}
const getstatename = (i: number) => {
  switch (i) {
    case 0:
      return "运行中";
    case 1:
      return "等待";
    case 2:
      return "暂停";
  }
  return "";
};
const api = new URL(document.URL);
if (import.meta.env.DEV) {
  api.host = "127.0.0.1:13232";
}
api.pathname += "/api/streamer";
const streamer = ref<Streamerlist>();
const create = ref<Streamer>({
  name: "",
  streamurl: "",
  sourceurl: "",
  autorestart: false,
  status: -1,
});
const opencreate = ref<boolean>(false);
const getStreamList = async () => {
  const response = await fetch(api.toString(), {
    method: "GET",
  });
  streamer.value = await response.json();
};
const Save = async (name: string, item: Streamer) => {
  if (item.name != name) {
    const response = await fetch(`${api.toString()}/${name}`, {
      method: "DELETE",
    });
  }
  const response = await fetch(api.toString(), {
    method: "POST",
    body: JSON.stringify(item),
    headers: {
      "Content-Type": "application/json",
    },
  });
  getStreamList();
};
const start = async (name: string) => {
  const response = await fetch(`${api.toString()}/${name}/start`, {
    method: "POST",
  });
  getStreamList();
};
const stop = async (name: string) => {
  const response = await fetch(`${api.toString()}/${name}/stop`, {
    method: "POST",
  });
  getStreamList();
};
const del = async (name: string) => {
  const response = await fetch(`${api.toString()}/${name}`, {
    method: "DELETE",
  });
  getStreamList();
};
getStreamList();
</script>

<style>
#app {
  color: var(--el-text-color-primary);
}
.header {
  display: flex;
  flex-direction: row;
}
</style>
