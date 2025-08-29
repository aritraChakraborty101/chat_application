import 'dart:io';
import 'package:dio/dio.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_secure_storage/flutter_secure_storage.dart';
import '../models/user.dart';

class ApiClient {
  static const String baseUrl = 'http://localhost:8080/api/v1';
  static const String tokenKey = 'auth_token';

  late final Dio _dio;
  final FlutterSecureStorage _storage = const FlutterSecureStorage();

  ApiClient() {
    _dio = Dio(
      BaseOptions(
        baseUrl: baseUrl,
        connectTimeout: const Duration(seconds: 30),
        receiveTimeout: const Duration(seconds: 30),
        headers: {'Content-Type': 'application/json'},
      ),
    );

    // Add request interceptor to include auth token
    _dio.interceptors.add(
      InterceptorsWrapper(
        onRequest: (options, handler) async {
          final token = await _storage.read(key: tokenKey);
          if (token != null) {
            options.headers['Authorization'] = 'Bearer $token';
          }
          handler.next(options);
        },
        onError: (error, handler) {
          // Handle 401 errors by clearing token
          if (error.response?.statusCode == 401) {
            _storage.delete(key: tokenKey);
          }
          handler.next(error);
        },
      ),
    );
  }

  // Auth methods
  Future<LoginResponse> register(RegisterRequest request) async {
    try {
      final response = await _dio.post(
        '/auth/register',
        data: request.toJson(),
      );
      final loginResponse = LoginResponse.fromJson(response.data);
      await _storage.write(key: tokenKey, value: loginResponse.token);
      return loginResponse;
    } on DioException catch (e) {
      throw _handleError(e);
    }
  }

  Future<LoginResponse> login(LoginRequest request) async {
    try {
      final response = await _dio.post('/auth/login', data: request.toJson());
      final loginResponse = LoginResponse.fromJson(response.data);
      await _storage.write(key: tokenKey, value: loginResponse.token);
      return loginResponse;
    } on DioException catch (e) {
      throw _handleError(e);
    }
  }

  Future<void> logout() async {
    await _storage.delete(key: tokenKey);
  }

  Future<String?> getToken() async {
    return await _storage.read(key: tokenKey);
  }

  Future<bool> isLoggedIn() async {
    final token = await getToken();
    return token != null;
  }

  // User methods
  Future<User> getCurrentUser() async {
    try {
      final response = await _dio.get('/users/me');
      return User.fromJson(response.data);
    } on DioException catch (e) {
      throw _handleError(e);
    }
  }

  Future<User> getUserById(String id) async {
    try {
      final response = await _dio.get('/users/$id');
      return User.fromJson(response.data);
    } on DioException catch (e) {
      throw _handleError(e);
    }
  }

  Future<void> updateProfile(UpdateProfileRequest request) async {
    try {
      await _dio.put('/users/me', data: request.toJson());
    } on DioException catch (e) {
      throw _handleError(e);
    }
  }

  Future<List<User>> searchUsers(String query, {int limit = 20}) async {
    try {
      final response = await _dio.get(
        '/users/search',
        queryParameters: {'q': query, 'limit': limit},
      );
      return (response.data as List)
          .map((json) => User.fromJson(json))
          .toList();
    } on DioException catch (e) {
      throw _handleError(e);
    }
  }

  // Connection methods
  Future<void> sendConnectionRequest(String addresseeId) async {
    try {
      await _dio.post('/connections/send-request/$addresseeId');
    } on DioException catch (e) {
      throw _handleError(e);
    }
  }

  Future<void> acceptConnectionRequest(String requesterId) async {
    try {
      await _dio.post('/connections/accept-request/$requesterId');
    } on DioException catch (e) {
      throw _handleError(e);
    }
  }

  Future<void> declineConnectionRequest(String requesterId) async {
    try {
      await _dio.post('/connections/decline-request/$requesterId');
    } on DioException catch (e) {
      throw _handleError(e);
    }
  }

  Future<void> removeConnection(String friendId) async {
    try {
      await _dio.delete('/connections/remove-friend/$friendId');
    } on DioException catch (e) {
      throw _handleError(e);
    }
  }

  Future<List<ConnectionWithUser>> getConnections() async {
    try {
      final response = await _dio.get('/connections');
      return (response.data as List)
          .map((json) => ConnectionWithUser.fromJson(json))
          .toList();
    } on DioException catch (e) {
      throw _handleError(e);
    }
  }

  Future<List<ConnectionWithUser>> getPendingRequests() async {
    try {
      final response = await _dio.get('/connections/pending');
      return (response.data as List)
          .map((json) => ConnectionWithUser.fromJson(json))
          .toList();
    } on DioException catch (e) {
      throw _handleError(e);
    }
  }

  // Error handling
  ApiException _handleError(DioException error) {
    if (error.response != null) {
      final data = error.response!.data;
      if (data is Map<String, dynamic>) {
        return ApiException(
          message: data['message'] ?? data['error'] ?? 'Unknown error',
          statusCode: error.response!.statusCode,
        );
      }
    }

    if (error.type == DioExceptionType.connectionTimeout ||
        error.type == DioExceptionType.receiveTimeout) {
      return ApiException(message: 'Connection timeout');
    }

    if (error.type == DioExceptionType.connectionError) {
      return ApiException(message: 'No internet connection');
    }

    return ApiException(message: error.message ?? 'Unknown error');
  }
}

class ApiException implements Exception {
  final String message;
  final int? statusCode;

  ApiException({required this.message, this.statusCode});

  @override
  String toString() => 'ApiException: $message';
}

// Provider for API client
final apiClientProvider = Provider<ApiClient>((ref) {
  return ApiClient();
});
