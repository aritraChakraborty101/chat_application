import 'package:json_annotation/json_annotation.dart';
import 'package:equatable/equatable.dart';

part 'user.g.dart';

@JsonSerializable()
class User extends Equatable {
  final String id;
  final String username;
  @JsonKey(name: 'display_name')
  final String displayName;
  final String email; // Back to required since backend now provides it
  @JsonKey(name: 'created_at')
  final DateTime createdAt;

  const User({
    required this.id,
    required this.username,
    required this.displayName,
    required this.email, // Required again
    required this.createdAt,
  });

  factory User.fromJson(Map<String, dynamic> json) => _$UserFromJson(json);
  Map<String, dynamic> toJson() => _$UserToJson(this);

  @override
  List<Object?> get props => [id, username, displayName, email, createdAt];

  User copyWith({
    String? id,
    String? username,
    String? displayName,
    String? email,
    DateTime? createdAt,
  }) {
    return User(
      id: id ?? this.id,
      username: username ?? this.username,
      displayName: displayName ?? this.displayName,
      email: email ?? this.email,
      createdAt: createdAt ?? this.createdAt,
    );
  }
}

@JsonSerializable()
class UserConnection extends Equatable {
  final String id;
  @JsonKey(name: 'requester_id')
  final String requesterId;
  @JsonKey(name: 'addressee_id')
  final String addresseeId;
  final String status;
  @JsonKey(name: 'created_at')
  final DateTime createdAt;
  @JsonKey(name: 'updated_at')
  final DateTime updatedAt;

  const UserConnection({
    required this.id,
    required this.requesterId,
    required this.addresseeId,
    required this.status,
    required this.createdAt,
    required this.updatedAt,
  });

  factory UserConnection.fromJson(Map<String, dynamic> json) =>
      _$UserConnectionFromJson(json);
  Map<String, dynamic> toJson() => _$UserConnectionToJson(this);

  @override
  List<Object?> get props => [
        id,
        requesterId,
        addresseeId,
        status,
        createdAt,
        updatedAt,
      ];
}

@JsonSerializable()
class ConnectionWithUser extends Equatable {
  final UserConnection connection;
  final User user;

  const ConnectionWithUser({required this.connection, required this.user});

  factory ConnectionWithUser.fromJson(Map<String, dynamic> json) =>
      _$ConnectionWithUserFromJson(json);
  Map<String, dynamic> toJson() => _$ConnectionWithUserToJson(this);

  @override
  List<Object?> get props => [connection, user];
}

// Request/Response DTOs
@JsonSerializable()
class RegisterRequest {
  final String username;
  @JsonKey(name: 'display_name')
  final String displayName;
  final String email;
  final String password;

  const RegisterRequest({
    required this.username,
    required this.displayName,
    required this.email,
    required this.password,
  });

  factory RegisterRequest.fromJson(Map<String, dynamic> json) =>
      _$RegisterRequestFromJson(json);
  Map<String, dynamic> toJson() => _$RegisterRequestToJson(this);
}

@JsonSerializable()
class LoginRequest {
  final String email;
  final String password;

  const LoginRequest({required this.email, required this.password});

  factory LoginRequest.fromJson(Map<String, dynamic> json) =>
      _$LoginRequestFromJson(json);
  Map<String, dynamic> toJson() => _$LoginRequestToJson(this);
}

@JsonSerializable()
class LoginResponse {
  final String token;
  final User user;

  const LoginResponse({required this.token, required this.user});

  factory LoginResponse.fromJson(Map<String, dynamic> json) =>
      _$LoginResponseFromJson(json);
  Map<String, dynamic> toJson() => _$LoginResponseToJson(this);
}

@JsonSerializable()
class UpdateProfileRequest {
  @JsonKey(name: 'display_name')
  final String displayName;

  const UpdateProfileRequest({required this.displayName});

  factory UpdateProfileRequest.fromJson(Map<String, dynamic> json) =>
      _$UpdateProfileRequestFromJson(json);
  Map<String, dynamic> toJson() => _$UpdateProfileRequestToJson(this);
}

@JsonSerializable(genericArgumentFactories: true)
class ApiResponse<T> {
  final String? message;
  final T? data;
  final String? error;

  const ApiResponse({this.message, this.data, this.error});

  factory ApiResponse.fromJson(
    Map<String, dynamic> json,
    T Function(Object? json) fromJsonT,
  ) =>
      _$ApiResponseFromJson(json, fromJsonT);
  Map<String, dynamic> toJson(Object Function(T value) toJsonT) =>
      _$ApiResponseToJson(this, toJsonT);
}
