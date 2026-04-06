<script setup lang="ts">
import { useCrudList } from '../composables/useCrudList'
import { listPhishlets, createPhishlet, updatePhishlet, deletePhishlet, type Phishlet } from '../api/client'
import PageHeader from '../components/PageHeader.vue'
import DeleteButton from '../components/DeleteButton.vue'
import PhishletForm from '../components/PhishletForm.vue'
import ErrorBanner from '../components/ErrorBanner.vue'
import EmptyState from '../components/EmptyState.vue'
import DataTable from '../components/DataTable.vue'
import DataTableRow from '../components/DataTableRow.vue'
import AddButton from '../components/AddButton.vue'

const { items: phishlets, showForm, editingId, error, form: yaml, openCreate, openEdit, closeForm, submit, remove } = useCrudList<Phishlet, string>(
  { list: listPhishlets, create: createPhishlet, update: updatePhishlet, remove: deletePhishlet },
  { emptyForm: () => '', toForm: (p) => p.yaml, confirmMessage: 'Delete this phishlet?' },
)
</script>

<template>
  <div>
    <PageHeader title="Phishlets" :subtitle="`${phishlets.length} phishlets`">
      <AddButton @click="openCreate">New Phishlet</AddButton>
    </PageHeader>

    <ErrorBanner v-model="error" />

    <PhishletForm v-if="showForm" v-model="yaml" :submit-label="editingId ? 'Save' : 'Create'"
      @submit="submit" @cancel="closeForm" />

    <EmptyState v-if="phishlets.length === 0 && !showForm" message="No phishlets. Add one to define phishing site configurations." />

    <DataTable v-else-if="phishlets.length > 0" :columns="[{ label: 'Name' }, { label: 'Created' }, { label: '', width: 'w-16' }]">
      <DataTableRow
        v-for="(phishlet, i) in phishlets"
        :key="phishlet.id"
        :index="i"
        clickable
        @click="openEdit(phishlet)"
      >
        <td class="px-4 py-2.5 text-primary">{{ phishlet.name }}</td>
        <td class="px-4 py-2.5 text-dim">{{ new Date(phishlet.created_at).toLocaleDateString() }}</td>
        <td class="px-4 py-2.5 text-right">
          <DeleteButton @click.stop="remove(phishlet.id)" />
        </td>
      </DataTableRow>
    </DataTable>
  </div>
</template>
