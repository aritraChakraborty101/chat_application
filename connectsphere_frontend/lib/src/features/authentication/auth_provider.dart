import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:riverpod_annotation/riverpod_annotation.dart';
import '../../core/api_client.dart';
import '../../models/user.dart';

part 'auth_provider.g.dart';

// Auth state class
class AuthState {
  final User? user;
  final bool isLoading;
  final String? error;

  const AuthState({this.user, this.isLoading = false, this.error});

  AuthState copyWith({User? user, bool? isLoading, String? error}) {
    return AuthState(
      user: user ?? this.user,
      isLoading: isLoading ?? this.isLoading,
      error: error,
    );
  }

  bool get isAuthenticated => user != null;
}

// Auth notifier
@riverpod
class AuthNotifier extends _$AuthNotifier {
  @override
  AuthState build() {
    _checkAuthStatus();
    return const AuthState();
  }

  Future<void> _checkAuthStatus() async {
    final apiClient = ref.read(apiClientProvider);
    final isLoggedIn = await apiClient.isLoggedIn();

    if (isLoggedIn) {
      try {
        final user = await apiClient.getCurrentUser();
        state = state.copyWith(user: user);
      } catch (e) {
        // Token might be invalid, clear it
        await apiClient.logout();
        state = state.copyWith(user: null);
      }
    }
  }

  Future<bool> register({
    required String username,
    required String displayName,
    required String email,
    required String password,
  }) async {
    state = state.copyWith(isLoading: true, error: null);

    try {
      final apiClient = ref.read(apiClientProvider);
      final request = RegisterRequest(
        username: username,
        displayName: displayName,
        email: email,
        password: password,
      );

      final response = await apiClient.register(request);
      state = state.copyWith(user: response.user, isLoading: false);
      return true;
    } catch (e) {
      state = state.copyWith(isLoading: false, error: e.toString());
      return false;
    }
  }

  Future<bool> login({required String email, required String password}) async {
    state = state.copyWith(isLoading: true, error: null);

    try {
      final apiClient = ref.read(apiClientProvider);
      final request = LoginRequest(email: email, password: password);

      final response = await apiClient.login(request);
      state = state.copyWith(user: response.user, isLoading: false);
      return true;
    } catch (e) {
      state = state.copyWith(isLoading: false, error: e.toString());
      return false;
    }
  }

  Future<void> logout() async {
    final apiClient = ref.read(apiClientProvider);
    await apiClient.logout();
    state = const AuthState();
  }

  Future<bool> updateProfile(String displayName) async {
    if (state.user == null) return false;

    state = state.copyWith(isLoading: true, error: null);

    try {
      final apiClient = ref.read(apiClientProvider);
      final request = UpdateProfileRequest(displayName: displayName);

      await apiClient.updateProfile(request);

      // Update local user state
      final updatedUser = state.user!.copyWith(displayName: displayName);
      state = state.copyWith(user: updatedUser, isLoading: false);
      return true;
    } catch (e) {
      state = state.copyWith(isLoading: false, error: e.toString());
      return false;
    }
  }

  void clearError() {
    state = state.copyWith(error: null);
  }
}
