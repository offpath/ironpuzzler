var app = angular.module('huntApp', ['ngCookies']);

app.controller('huntCtrl', ['$scope', '$cookies', '$http', function ($scope, $cookies, $http) {
	    $scope.refreshHunt = function() {
		$http.get("/api/hunt?hunt_id=" + huntId).success(function (response) {
			$scope.info = response;
			$scope.signedIn = ($scope.info.Teams.CurrentTeam != "");
			if ($scope.info.Teams.BadSignIn) {
			    window.alert("Bad password");
			    $scope.logout();
			}
			if ($scope.info.Leaderboard.Token != "") {
			    $scope.channel = new goog.appengine.Channel($scope.info.Leaderboard.Token);
			    $scope.socket = $scope.channel.open();
			    $scope.socket.onmessage = $scope.onMessage;
			    $http.get("/api/leaderboard?hunt_id=" + huntId).success($scope.updateLeaderboard);
			}
		    });
	    }
	    
	    $scope.updateLeaderboard = function(response) {
		for (var i = 0; i < $scope.info.Leaderboard.Progress.length; i++) {
		    var id = $scope.info.Leaderboard.Progress[i].ID;
		    if (id in response) {
			$scope.info.Leaderboard.Progress[i].Updatable = response[id];
		    }
		}
	    }
	    
	    $scope.onMessage = function(message) {
		if (message == "refresh") {
		    $scope.refreshHunt();
		} else {
		    $http.get("/api/leaderboard?hunt_id=" + huntId +
			      "&puzzleid=" + message).success($scope.updateLeaderboard);
		}
	    }
	    
	    $scope.submitAnswer = function(pid, answer) {
		$http.get("/api/submitanswer?hunt_id=" + huntId +
			  "&puzzleid=" + pid +
			  "&answer=" + encodeURIComponent(answer)). success(function (response) {
				  window.alert(response);
			      });
	    }
	    
	    $scope.login = function() {
		$cookies.put("team_id", $scope.teamID);
		$cookies.put("password", $scope.password);
		$scope.teamID = "";
		$scope.password = "";
		$scope.refreshHunt();
	    }
	    
	    $scope.logout = function() {
		$cookies.remove("team_id");
		$cookies.remove("password");
		$scope.refreshHunt();
	    }
	    
	    $scope.updatePuzzles = function() {
		for (var i = 0; i < $scope.info.Puzzles.Puzzles.length; i++) {
		    var p = $scope.info.Puzzles.Puzzles[i];
		    $http.get("/api/updatepuzzle?hunt_id=" + huntId +
			      "&puzzleid=" + p.ID +
			      "&name=" + encodeURIComponent(p.Name) +
			      "&answer=" + encodeURIComponent(p.Answer));
		}
	    }
	    
	    $scope.puzzleFormat = {
		false: "Non-paper",
		true: "Paper",
	    }

	    $scope.signedIn = false;
	    $scope.teamID = "";
	    $scope.password = "";
	    $scope.refreshHunt();
	}]);

app.controller('leaderboardCtrl', ['$scope', '$cookies', '$http', function ($scope, $cookies, $http) {
	    $scope.test = function() {}
	    
	}]);