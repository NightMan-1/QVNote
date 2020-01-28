<template>
    <div class="scrooll-wrap">
            <nav class="sidebar-nav">
                <ul class="nav" :class="{'d-none':sidebarType !== 'notebooksList'}">
                    <li class="nav-title text-success">
                        {{$t('general.sidebarLibrary')}}
                    </li>
                    <li class="nav-item">
                        <button class="nav-link"
                                :class="{ 'active': currentNotebookID === 'Inbox' }"
                                @click="$router.push('/notes/Inbox', () => {})">
                            <span class="badge badge-primary">{{notesCountInbox}}</span>
                            <i class="fas fa-inbox mr-1"></i> {{$t('general.sidebarInbox')}}
                        </button>
                    </li>
                    <li class="nav-item">
                        <button class="nav-link"
                                :class="{ 'active': currentNotebookID === 'Favorites' }"
                                @click="$router.push('/notes/Favorites', () => {})">
                            <span class="badge badge-primary">{{notesCountFavorites}}</span>
                            <i class="fas fa-star mr-1"></i> {{$t('general.sidebarFavorites')}}
                        </button>
                    </li>
                    <li class="nav-item">
                        <button class="nav-link"
                                :class="{ 'active': currentNotebookID === 'Trash' }"
                                @click="$router.push('/notes/Trash', () => {})">
                            <span class="badge badge-primary">{{notesCountTrash}}</span>
                            <i class="fas fa-trash-alt mr-1"></i> {{$t('general.sidebarTrash')}}
                        </button>
                    </li>
                    <li class="nav-item">
                        <button class="nav-link"
                                :class="{ 'active': currentNotebookID === 'Allnotes' }"
                                @click="$router.push('/notes/Allnotes', () => {})">
                            <span class="badge badge-primary" v-if="notesCountTotal > 0">{{notesCountTotal}}</span>
                            <i class="fas fa-archive mr-1"></i> {{$t('general.sidebarAllNotes')}}
                        </button>
                    </li>

                    <li class="nav-title text-primary">{{$t('general.sidebarNotebooks')}}</li>
                    <li
                        class="nav-item"
                        v-for="item in notebooksList" v-if="item.name !== 'Inbox' && item.name !== 'Trash'"
                        :key="item.uuid"
                    >
                        <button class="nav-link nav-link-notebook"
                                @click="$router.push('/notes/' + item.uuid, () => {})"
                                :class="{ 'active': item.uuid === currentNotebookID }">
                            <span class="badge badge-primary">{{item.notesCount}}</span>
                                {{item.name}}
                        </button>
                    </li>
                </ul>

                <ul class="nav" :class="{'d-none':sidebarType !== 'tagsList'}">
                    <li class="nav-title text-primary">{{$t('general.sidebarTags')}}</li>
                    <li
                        class="nav-item"
                        v-for="item in tagsList"
                        :key="item.url"
                    >
                        <button class="nav-link nav-link-notebook bg-primary- border-0 w-100 text-left"
                                @click="$router.push('/tags/'+item.url, () => {})"
                                :class="{ 'active': item.name === currentTagURL }">
                            <span class="badge badge-primary">{{item.count}}</span>
                            {{item.name}}
                        </button>
                    </li>

                </ul>
            </nav>
    </div>
</template>

<script>
import mixin from './mixins'

export default {
    name: 'qvSidebar',
    mixins: [mixin]
}
</script>
