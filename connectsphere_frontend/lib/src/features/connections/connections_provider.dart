import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:riverpod_annotation/riverpod_annotation.dart';
import '../../core/api_client.dart';
import '../../models/user.dart';

part 'connections_provider.g.dart';

// Connections state
class ConnectionsState {
  final List<ConnectionWithUser> connections;
  final List<ConnectionWithUser> pendingRequests;
  final List<User> searchResults;
  final bool isLoading;
  final String? error;

  const ConnectionsState({
    this.connections = const [],
    this.pendingRequests = const [],
    this.searchResults = const [],
    this.isLoading = false,
    this.error,
  });

  ConnectionsState copyWith({
    List<ConnectionWithUser>? connections,
    List<ConnectionWithUser>? pendingRequests,
    List<User>? searchResults,
    bool? isLoading,
    String? error,
  }) {
    return ConnectionsState(
      connections: connections ?? this.connections,
      pendingRequests: pendingRequests ?? this.pendingRequests,
      searchResults: searchResults ?? this.searchResults,
      isLoading: isLoading ?? this.isLoading,
      error: error,
    );
  }
}

@riverpod
class ConnectionsNotifier extends _$ConnectionsNotifier {
  @override
  ConnectionsState build() {
    return const ConnectionsState();
  }

  Future<void> loadConnections() async {
    state = state.copyWith(isLoading: true, error: null);

    try {
      final apiClient = ref.read(apiClientProvider);
      final connections = await apiClient.getConnections();

      state = state.copyWith(connections: connections, isLoading: false);
    } catch (e) {
      state = state.copyWith(isLoading: false, error: e.toString());
    }
  }

  Future<void> loadPendingRequests() async {
    state = state.copyWith(isLoading: true, error: null);

    try {
      final apiClient = ref.read(apiClientProvider);
      final requests = await apiClient.getPendingRequests();

      state = state.copyWith(pendingRequests: requests, isLoading: false);
    } catch (e) {
      state = state.copyWith(isLoading: false, error: e.toString());
    }
  }

  Future<void> searchUsers(String query) async {
    if (query.trim().isEmpty) {
      state = state.copyWith(searchResults: []);
      return;
    }

    state = state.copyWith(isLoading: true, error: null);

    try {
      final apiClient = ref.read(apiClientProvider);
      final results = await apiClient.searchUsers(query);

      state = state.copyWith(searchResults: results, isLoading: false);
    } catch (e) {
      state = state.copyWith(isLoading: false, error: e.toString());
    }
  }

  Future<bool> sendConnectionRequest(String addresseeId) async {
    try {
      final apiClient = ref.read(apiClientProvider);
      await apiClient.sendConnectionRequest(addresseeId);
      return true;
    } catch (e) {
      state = state.copyWith(error: e.toString());
      return false;
    }
  }

  Future<bool> acceptConnectionRequest(String requesterId) async {
    try {
      final apiClient = ref.read(apiClientProvider);
      await apiClient.acceptConnectionRequest(requesterId);

      // Refresh pending requests and connections
      await loadPendingRequests();
      await loadConnections();

      return true;
    } catch (e) {
      state = state.copyWith(error: e.toString());
      return false;
    }
  }

  Future<bool> declineConnectionRequest(String requesterId) async {
    try {
      final apiClient = ref.read(apiClientProvider);
      await apiClient.declineConnectionRequest(requesterId);

      // Refresh pending requests
      await loadPendingRequests();

      return true;
    } catch (e) {
      state = state.copyWith(error: e.toString());
      return false;
    }
  }

  Future<bool> removeConnection(String friendId) async {
    try {
      final apiClient = ref.read(apiClientProvider);
      await apiClient.removeConnection(friendId);

      // Refresh connections
      await loadConnections();

      return true;
    } catch (e) {
      state = state.copyWith(error: e.toString());
      return false;
    }
  }

  void clearError() {
    state = state.copyWith(error: null);
  }

  void clearSearchResults() {
    state = state.copyWith(searchResults: []);
  }
}
