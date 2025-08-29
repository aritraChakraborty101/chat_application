import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'connections_provider.dart';
import '../../models/user.dart';

class ConnectionsScreen extends ConsumerStatefulWidget {
  const ConnectionsScreen({super.key});

  @override
  ConsumerState<ConnectionsScreen> createState() => _ConnectionsScreenState();
}

class _ConnectionsScreenState extends ConsumerState<ConnectionsScreen>
    with SingleTickerProviderStateMixin {
  late TabController _tabController;

  @override
  void initState() {
    super.initState();
    _tabController = TabController(length: 2, vsync: this);

    // Load initial data
    WidgetsBinding.instance.addPostFrameCallback((_) {
      ref.read(connectionsNotifierProvider.notifier).loadConnections();
      ref.read(connectionsNotifierProvider.notifier).loadPendingRequests();
    });
  }

  @override
  void dispose() {
    _tabController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final connectionsState = ref.watch(connectionsNotifierProvider);

    return Scaffold(
      appBar: AppBar(
        title: const Text('Connections'),
        centerTitle: true,
        backgroundColor: Theme.of(context).primaryColor,
        foregroundColor: Colors.white,
        bottom: TabBar(
          controller: _tabController,
          indicatorColor: Colors.white,
          labelColor: Colors.white,
          unselectedLabelColor: Colors.white70,
          tabs: [
            Tab(text: 'Friends (${connectionsState.connections.length})'),
            Tab(text: 'Requests (${connectionsState.pendingRequests.length})'),
          ],
        ),
      ),
      body: TabBarView(
        controller: _tabController,
        children: [
          _buildConnectionsList(connectionsState),
          _buildPendingRequestsList(connectionsState),
        ],
      ),
    );
  }

  Widget _buildConnectionsList(ConnectionsState state) {
    if (state.isLoading && state.connections.isEmpty) {
      return const Center(child: CircularProgressIndicator());
    }

    if (state.connections.isEmpty) {
      return const Center(
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            Icon(Icons.people_outline, size: 64, color: Colors.grey),
            SizedBox(height: 16),
            Text(
              'No friends yet',
              style: TextStyle(fontSize: 18, color: Colors.grey),
            ),
            SizedBox(height: 8),
            Text(
              'Start connecting with people!',
              style: TextStyle(color: Colors.grey),
            ),
          ],
        ),
      );
    }

    return RefreshIndicator(
      onRefresh: () =>
          ref.read(connectionsNotifierProvider.notifier).loadConnections(),
      child: ListView.builder(
        itemCount: state.connections.length,
        itemBuilder: (context, index) {
          final connection = state.connections[index];
          return _buildConnectionTile(connection);
        },
      ),
    );
  }

  Widget _buildPendingRequestsList(ConnectionsState state) {
    if (state.isLoading && state.pendingRequests.isEmpty) {
      return const Center(child: CircularProgressIndicator());
    }

    if (state.pendingRequests.isEmpty) {
      return const Center(
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            Icon(Icons.inbox_outlined, size: 64, color: Colors.grey),
            SizedBox(height: 16),
            Text(
              'No pending requests',
              style: TextStyle(fontSize: 18, color: Colors.grey),
            ),
          ],
        ),
      );
    }

    return RefreshIndicator(
      onRefresh: () =>
          ref.read(connectionsNotifierProvider.notifier).loadPendingRequests(),
      child: ListView.builder(
        itemCount: state.pendingRequests.length,
        itemBuilder: (context, index) {
          final request = state.pendingRequests[index];
          return _buildPendingRequestTile(request);
        },
      ),
    );
  }

  Widget _buildConnectionTile(ConnectionWithUser connection) {
    return ListTile(
      leading: CircleAvatar(
        backgroundColor: Theme.of(context).primaryColor,
        child: Text(
          connection.user.displayName[0].toUpperCase(),
          style: const TextStyle(
            color: Colors.white,
            fontWeight: FontWeight.bold,
          ),
        ),
      ),
      title: Text(
        connection.user.displayName,
        style: const TextStyle(fontWeight: FontWeight.w500),
      ),
      subtitle: Text('@${connection.user.username}'),
      trailing: PopupMenuButton<String>(
        onSelected: (value) async {
          if (value == 'remove') {
            final confirmed = await _showConfirmDialog(
              context,
              'Remove Friend',
              'Are you sure you want to remove ${connection.user.displayName} from your friends?',
            );
            if (confirmed) {
              ref
                  .read(connectionsNotifierProvider.notifier)
                  .removeConnection(connection.user.id);
            }
          }
        },
        itemBuilder: (context) => [
          const PopupMenuItem(
            value: 'remove',
            child: Row(
              children: [
                Icon(Icons.person_remove, color: Colors.red),
                SizedBox(width: 8),
                Text('Remove Friend'),
              ],
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildPendingRequestTile(ConnectionWithUser request) {
    return ListTile(
      leading: CircleAvatar(
        backgroundColor: Theme.of(context).primaryColor,
        child: Text(
          request.user.displayName[0].toUpperCase(),
          style: const TextStyle(
            color: Colors.white,
            fontWeight: FontWeight.bold,
          ),
        ),
      ),
      title: Text(
        request.user.displayName,
        style: const TextStyle(fontWeight: FontWeight.w500),
      ),
      subtitle: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text('@${request.user.username}'),
          const SizedBox(height: 4),
          const Text(
            'Wants to connect with you',
            style: TextStyle(fontStyle: FontStyle.italic, color: Colors.grey),
          ),
        ],
      ),
      trailing: Row(
        mainAxisSize: MainAxisSize.min,
        children: [
          IconButton(
            icon: const Icon(Icons.check, color: Colors.green),
            onPressed: () {
              ref
                  .read(connectionsNotifierProvider.notifier)
                  .acceptConnectionRequest(request.user.id);
            },
          ),
          IconButton(
            icon: const Icon(Icons.close, color: Colors.red),
            onPressed: () {
              ref
                  .read(connectionsNotifierProvider.notifier)
                  .declineConnectionRequest(request.user.id);
            },
          ),
        ],
      ),
    );
  }

  Future<bool> _showConfirmDialog(
    BuildContext context,
    String title,
    String content,
  ) async {
    return await showDialog<bool>(
          context: context,
          builder: (context) => AlertDialog(
            title: Text(title),
            content: Text(content),
            actions: [
              TextButton(
                onPressed: () => Navigator.of(context).pop(false),
                child: const Text('Cancel'),
              ),
              TextButton(
                onPressed: () => Navigator.of(context).pop(true),
                child: const Text('Confirm'),
              ),
            ],
          ),
        ) ??
        false;
  }
}
